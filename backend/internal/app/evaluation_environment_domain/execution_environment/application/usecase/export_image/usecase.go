package exportimage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/bundle"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/fingerprint"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/app/shared/types/zipfile"
)

var (
	ErrorImageNotFound  = errors.New("image not found")
	ErrorInvalidMode    = errors.New("mode must be source or archive")
	ErrorInternalServer = errors.New("internal server error")
	// ErrorNotBuilt is returned when an archive is asked for an image the
	// daemon has never built. Exporting an empty tar would produce a bundle
	// that imports cleanly and leaves the venue with nothing to judge against.
	ErrorNotBuilt = errors.New("此映像尚未建置完成，無法以 archive 模式匯出；請先建置，或改用 source 模式")
)

type Input struct {
	Tag  string
	Mode string
}

type Output struct {
	Filename string
	Content  []byte
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// ImageArchiver is the daemon-side half of an archive export. Declared here so
// the whole usecase can be tested without a Docker daemon.
type ImageArchiver interface {
	// Save streams the image out, as docker save would.
	Save(ctx context.Context, tag string) (io.ReadCloser, error)
	// Digest is the content identity the daemon reports for a built image.
	Digest(ctx context.Context, tag string) (string, error)
}

type usecase struct {
	repository repository.ImageRepository
	fs         sharedservice.FileSystemService
	zip        sharedservice.ZipService
	archiver   ImageArchiver
}

func NewUsecase(
	repository repository.ImageRepository,
	fs sharedservice.FileSystemService,
	zip sharedservice.ZipService,
	archiver ImageArchiver,
) Usecase {
	return &usecase{repository: repository, fs: fs, zip: zip, archiver: archiver}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if input.Mode != bundle.ModeSource && input.Mode != bundle.ModeArchive {
		return nil, ErrorInvalidMode
	}

	img, err := u.repository.FindByTag(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}
	if img == nil {
		return nil, ErrorImageNotFound
	}

	definition := &bundle.Image{
		FormatVersion: bundle.FormatVersion,
		Kind:          bundle.KindImage,
		Tag:           img.Tag().String(),
		MountPath:     img.MountPath().String(),
		Description:   img.Description(),
		Mode:          input.Mode,
	}

	var payload zipfile.File
	if input.Mode == bundle.ModeArchive {
		payload, definition.Fingerprint, err = u.archive(img.Tag().String(), img.IsReady())
	} else {
		payload, definition.Fingerprint, err = u.source(img.Project().StoragePath())
	}
	if err != nil {
		return nil, err
	}

	definitionJSON, err := bundle.Marshal(definition)
	if err != nil {
		return nil, ErrorInternalServer
	}

	files := []zipfile.File{
		payload,
		{Path: bundle.ImageFile, Body: string(definitionJSON)},
	}

	archive, err := u.zip.ZipFromFiles(&files)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Filename: filenameFor(input.Tag, input.Mode),
		Content:  archive.Bytes(),
	}, nil
}

// archive asks the daemon for the built image.
func (u *usecase) archive(tag string, isReady bool) (zipfile.File, string, error) {
	// Checked before touching the daemon: an unbuilt image may still have a
	// row and a build context, and docker save on a tag it does not hold
	// fails in a way that reads like an infrastructure fault rather than the
	// operator's answerable mistake.
	if !isReady {
		return zipfile.File{}, "", ErrorNotBuilt
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	digest, err := u.archiver.Digest(ctx, tag)
	if err != nil {
		return zipfile.File{}, "", ErrorNotBuilt
	}

	stream, err := u.archiver.Save(ctx, tag)
	if err != nil {
		return zipfile.File{}, "", ErrorInternalServer
	}
	defer stream.Close()

	content, err := io.ReadAll(stream)
	if err != nil {
		return zipfile.File{}, "", ErrorInternalServer
	}

	return zipfile.File{Path: bundle.ArchiveFile, Body: string(content)}, fingerprint.OfDigest(digest), nil
}

// source packages the build context the image was created from.
func (u *usecase) source(contextPath string) (zipfile.File, string, error) {
	if !u.fs.Exists(contextPath) {
		return zipfile.File{}, "", fmt.Errorf("%w: 這個映像沒有保留建置原始檔，請改用 archive 模式", ErrorInternalServer)
	}

	print, err := fingerprint.OfBuildContext(u.fs, contextPath)
	if err != nil {
		return zipfile.File{}, "", ErrorInternalServer
	}

	zipped, err := u.zip.ZipFromPath(contextPath)
	if err != nil {
		return zipfile.File{}, "", ErrorInternalServer
	}

	return zipfile.File{Path: bundle.ContextFile, Body: zipped.String()}, print, nil
}

// filenameFor turns a tag into something a filesystem accepts: a tag holds a
// colon and may hold slashes, neither of which survives a download intact on
// every platform.
func filenameFor(tag, mode string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		default:
			return '-'
		}
	}, tag)
	return fmt.Sprintf("image-%s-%s.zip", strings.Trim(safe, "-"), mode)
}
