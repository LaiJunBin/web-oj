package importimage

import (
	archivetar "archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/bundle"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/fingerprint"
	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
	updateimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/update_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	sharedservice "judge-backend/internal/app/shared/service"
)

var (
	ErrorInternalServer = errors.New("internal server error")
	ErrorInvalidBundle  = errors.New("這不是有效的映像匯出檔")
	ErrorConflict       = errors.New("此 tag 已存在且內容不同")
)

// InvalidBundleError reports every fault at once, so a hand-written bundle
// does not need one upload per typo.
type InvalidBundleError struct{ Details []string }

func (e *InvalidBundleError) Error() string { return ErrorInvalidBundle.Error() }
func (e *InvalidBundleError) Unwrap() error { return ErrorInvalidBundle }

// ConflictError is the tag existing with different content.
//
// Separate from InvalidBundleError because the operator's next move differs:
// a malformed bundle has to be fixed, a conflict has to be decided. Details
// name what would be overwritten and who is using it, because "force" is only
// an informed choice if the blast radius is visible.
type ConflictError struct{ Details []string }

func (e *ConflictError) Error() string { return ErrorConflict.Error() }
func (e *ConflictError) Unwrap() error { return ErrorConflict }

type Input struct {
	Bundle *bytes.Buffer
	// Force overwrites an existing tag whose content differs. Without it a
	// conflict is refused, because the tag may be what several live contests
	// are judging against.
	Force bool
}

type Output struct {
	Tag string
	// Mode says which shape arrived, and BuildQueued whether the operator
	// still has to wait: a source bundle has to be built, an archive is
	// usable the moment it loads.
	Mode        string
	BuildQueued bool
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// ImageLoader is the daemon-side half of an archive import.
type ImageLoader interface {
	Load(ctx context.Context, tar io.Reader) error
	Digest(ctx context.Context, tag string) (string, error)
}

// TagUsage reports which projects reference a tag, so a conflict can say what
// forcing would affect.
type TagUsage interface {
	ProjectsUsing(tag string) ([]string, error)
}

type usecase struct {
	repository  repository.ImageRepository
	factory     *factory.ImageFactory
	fs          sharedservice.FileSystemService
	zip         sharedservice.ZipService
	loader      ImageLoader
	usage       TagUsage
	createImage createimage.Usecase
	updateImage updateimage.Usecase
}

func NewUsecase(
	repository repository.ImageRepository,
	factory *factory.ImageFactory,
	fs sharedservice.FileSystemService,
	zip sharedservice.ZipService,
	loader ImageLoader,
	usage TagUsage,
	createImage createimage.Usecase,
	updateImage updateimage.Usecase,
) Usecase {
	return &usecase{
		repository: repository, factory: factory, fs: fs, zip: zip,
		loader: loader, usage: usage, createImage: createImage, updateImage: updateImage,
	}
}

// Execute reads the bundle, decides what it would do to what is already here,
// and only then writes — the same order the project import uses, and for the
// same reason: a bundle that is going to be refused must not have touched
// anything first.
func (u *usecase) Execute(input *Input) (*Output, error) {
	definition, payload, err := u.verify(input)
	if err != nil {
		return nil, err
	}

	existing, err := u.repository.FindByTag(definition.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if existing != nil {
		same, err := u.sameContent(definition, existing.Project().StoragePath())
		if err != nil {
			return nil, err
		}

		// Already here, byte for byte. Silently reusing is right: this is the
		// ordinary case when a venue imports two problems that share an image.
		if same {
			return &Output{Tag: definition.Tag, Mode: definition.Mode, BuildQueued: false}, nil
		}

		if !input.Force {
			return nil, u.conflict(definition.Tag)
		}

		// Forcing replaces the image IN PLACE — the same thing an operator
		// does by hand with PATCH and a new context — rather than deleting
		// and recreating the row. The first version deleted, and that broke
		// twice over: the soft-deleted twin took over tag resolution, so the
		// replacement's builds attached to the dead row; and on MySQL, which
		// has no partial indexes, the "unique among live rows" index is
		// simply unique, so recreating the tag failed outright. Keeping the
		// row keeps the id, and everything downstream — build history, the
		// builds list, rebuild — keeps working because nothing moved.
		// (The USER delete path solves the same MySQL constraint differently:
		// it renames the soft-deleted row's tag — model.DeletedImageTag — so
		// the tag frees up for re-creation. In-place replacement here remains
		// the right call for force import precisely because it keeps the id.)
		return u.replaceExisting(definition, payload)
	}

	if definition.Mode == bundle.ModeArchive {
		return u.loadArchive(definition, payload)
	}
	return u.buildFromSource(definition, payload)
}

// replaceExisting overwrites a live image's content under its own id.
func (u *usecase) replaceExisting(definition *bundle.Image, payload []byte) (*Output, error) {
	if definition.Mode == bundle.ModeArchive {
		return u.loadArchiveOverExisting(definition, payload)
	}

	// Source mode is exactly the admin's own "update with a new context":
	// fields, project setup and the queued rebuild all through the usecase
	// that already owns that flow.
	mountPath := definition.MountPath
	description := definition.Description
	if _, err := u.updateImage.Execute(&updateimage.Input{
		Tag:            definition.Tag,
		MountPath:      &mountPath,
		Description:    &description,
		ProjectZipFile: bytes.NewBuffer(payload),
		NoCache:        true,
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorInternalServer, err.Error())
	}

	return &Output{Tag: definition.Tag, Mode: definition.Mode, BuildQueued: true}, nil
}

func (u *usecase) verify(input *Input) (*bundle.Image, []byte, error) {
	if input.Bundle == nil || input.Bundle.Len() == 0 {
		return nil, nil, &InvalidBundleError{Details: []string{"沒有收到檔案"}}
	}
	// ReadZipToMap bounds the archive's total size and entry count, so an
	// image bundle (whose archive mode can legitimately be hundreds of MB)
	// still cannot be a zip bomb expanded straight into memory here.
	files, err := u.zip.ReadZipToMap(input.Bundle)
	if err != nil {
		return nil, nil, &InvalidBundleError{Details: []string{"上傳的檔案不是有效或過大的 zip"}}
	}

	definitionJSON, ok := files[bundle.ImageFile]
	if !ok {
		return nil, nil, &InvalidBundleError{Details: []string{"匯出檔裡沒有 " + bundle.ImageFile}}
	}

	definition, err := bundle.Unmarshal(definitionJSON)
	if err != nil {
		return nil, nil, &InvalidBundleError{Details: []string{bundle.ImageFile + " 無法解析：" + err.Error()}}
	}

	details := make([]string, 0)
	for _, p := range definition.Validate() {
		details = append(details, p.Error())
	}

	payload, ok := files[definition.PayloadFile()]
	if !ok {
		details = append(details, fmt.Sprintf("mode 為 %q，但匯出檔裡沒有 %s", definition.Mode, definition.PayloadFile()))
	} else if definition.Mode == bundle.ModeSource {
		if err := u.zip.CheckZip(bytes.NewBuffer(payload)); err != nil {
			details = append(details, bundle.ContextFile+" 不是有效的 zip")
		}
	}

	if len(details) > 0 {
		return nil, nil, &InvalidBundleError{Details: details}
	}
	return definition, payload, nil
}

// sameContent compares the bundle against what is already installed.
//
// Only a source bundle can be compared: its fingerprint covers the build
// context, which this installation still has on disk. An archive's fingerprint
// is the digest of layers built elsewhere, and the same context built twice
// does not produce the same digest — so an archive whose tag already exists is
// always treated as different, and needs an explicit decision.
func (u *usecase) sameContent(definition *bundle.Image, existingContext string) (bool, error) {
	if definition.Mode != bundle.ModeSource || !u.fs.Exists(existingContext) {
		return false, nil
	}

	print, err := fingerprint.OfBuildContext(u.fs, existingContext)
	if err != nil {
		return false, ErrorInternalServer
	}
	return print == definition.Fingerprint, nil
}

// conflict explains what forcing would cost.
func (u *usecase) conflict(tag string) error {
	details := []string{
		fmt.Sprintf("映像 %q 已存在，但內容與匯出檔不同", tag),
	}

	users, err := u.usage.ProjectsUsing(tag)
	if err == nil && len(users) > 0 {
		details = append(details, fmt.Sprintf("目前有 %d 個題目正在使用它：%v", len(users), users))
		details = append(details, "強制覆蓋會一併改變這些題目的評測環境")
	}
	details = append(details, "確定要取代請重新送出並勾選強制覆蓋")

	return &ConflictError{Details: details}
}

// buildFromSource hands the context to the ordinary creation path, which
// reserves a build slot and queues the build. Nothing here waits for it: the
// build has its own page, its own status and its own log stream, and an import
// that blocked for minutes would only be a worse version of them.
func (u *usecase) buildFromSource(definition *bundle.Image, context []byte) (*Output, error) {
	_, err := u.createImage.Execute(&createimage.Input{
		Tag:            definition.Tag,
		MountPath:      definition.MountPath,
		Description:    definition.Description,
		ProjectZipFile: bytes.NewBuffer(context),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorInternalServer, err.Error())
	}

	return &Output{Tag: definition.Tag, Mode: definition.Mode, BuildQueued: true}, nil
}

// loadArchiveOverExisting is the archive half of a forced replacement: the
// daemon takes the new layers, and the SAME row is updated and marked ready —
// no second row, no soft-deleted twin.
func (u *usecase) loadArchiveOverExisting(definition *bundle.Image, tar []byte) (*Output, error) {
	if err := u.loadIntoDaemon(definition, tar); err != nil {
		return nil, err
	}

	existing, err := u.repository.FindByTag(definition.Tag)
	if err != nil || existing == nil {
		return nil, ErrorInternalServer
	}

	if err := existing.SetMountPath(definition.MountPath); err != nil {
		return nil, ErrorInternalServer
	}
	existing.SetDescription(definition.Description)
	existing.MarkAsReady()

	if err := u.repository.Save(existing); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{Tag: definition.Tag, Mode: definition.Mode, BuildQueued: false}, nil
}

func (u *usecase) loadIntoDaemon(definition *bundle.Image, tar []byte) error {
	// Reject an archive carrying any tag other than the one being imported
	// BEFORE handing it to the daemon: docker load applies EVERY RepoTag in the
	// archive, so an extra tag could silently overwrite another in-use judge
	// image. This is the archive-mode counterpart of the digest check below.
	if err := verifyArchiveContainsOnlyTag(tar, definition.Tag); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := u.loader.Load(ctx, bytes.NewReader(tar)); err != nil {
		return fmt.Errorf("%w: 載入映像失敗", ErrorInternalServer)
	}

	// Confirmed against the daemon rather than assumed: docker load reports
	// success for an archive that contains no matching tag, and a row marked
	// ready for an image the daemon does not have would fail at judging time
	// instead of here.
	if _, err := u.loader.Digest(ctx, definition.Tag); err != nil {
		return fmt.Errorf("%w: 載入後在 daemon 上找不到 %q，匯出檔可能不完整", ErrorInternalServer, definition.Tag)
	}

	return nil
}

// verifyArchiveContainsOnlyTag parses a docker-save archive's manifest and
// rejects it unless it is a single image whose every RepoTag is exactly
// allowedTag. Without this a bundle could smuggle extra tags into the daemon —
// docker load applies them all — overwriting another in-use judge image, or
// bundle additional images entirely.
func verifyArchiveContainsOnlyTag(archive []byte, allowedTag string) error {
	tr := archivetar.NewReader(bytes.NewReader(archive))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return &InvalidBundleError{Details: []string{"映像封存檔無法讀取"}}
		}
		if path.Clean(hdr.Name) != "manifest.json" {
			continue
		}

		var entries []struct {
			RepoTags []string `json:"RepoTags"`
		}
		if err := json.NewDecoder(tr).Decode(&entries); err != nil {
			return &InvalidBundleError{Details: []string{"映像封存檔的 manifest.json 無法解析"}}
		}
		if len(entries) != 1 {
			return &InvalidBundleError{Details: []string{fmt.Sprintf("映像封存檔必須只含一個映像，卻含有 %d 個", len(entries))}}
		}
		for _, tag := range entries[0].RepoTags {
			if tag != allowedTag {
				return &InvalidBundleError{Details: []string{fmt.Sprintf("映像封存檔含有未預期的 tag %q（只允許 %q）", tag, allowedTag)}}
			}
		}
		return nil
	}
	return &InvalidBundleError{Details: []string{"映像封存檔缺少 manifest.json"}}
}

// loadArchive puts prebuilt layers straight into the daemon. No build, no
// network — this is the shape an offline venue needs.
func (u *usecase) loadArchive(definition *bundle.Image, tar []byte) (*Output, error) {
	if err := u.loadIntoDaemon(definition, tar); err != nil {
		return nil, err
	}

	img, err := u.factory.CreateImage(definition.Tag, definition.MountPath, definition.Description)
	if err != nil {
		return nil, ErrorInternalServer
	}
	img.MarkAsReady()

	if err := u.repository.Save(img); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{Tag: definition.Tag, Mode: definition.Mode, BuildQueued: false}, nil
}
