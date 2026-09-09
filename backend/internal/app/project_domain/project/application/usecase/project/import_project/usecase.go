package importproject

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"judge-backend/internal/app/project_domain/project/application/bundle"
	creategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/create_grading_item"
	updategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/update_grading_item"
	createlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/create_lifecycle"
	createproject "judge-backend/internal/app/project_domain/project/application/usecase/project/create_project"
	deleteproject "judge-backend/internal/app/project_domain/project/application/usecase/project/delete_project"
	createtestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/create_testcase"
	updatetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/update_testcase"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	domainservice "judge-backend/internal/app/project_domain/project/domain/service"
	sharedservice "judge-backend/internal/app/shared/service"
)

var (
	ErrorInternalServer = errors.New("internal server error")
	// ErrorInvalidBundle carries the faults found before anything was written.
	// Its Details are what the operator has to fix.
	ErrorInvalidBundle = errors.New("這不是有效的題目匯出檔")
)

// InvalidBundleError reports every fault in one go, each located by its path
// in project.json.
type InvalidBundleError struct {
	Details []string
}

func (e *InvalidBundleError) Error() string { return ErrorInvalidBundle.Error() }
func (e *InvalidBundleError) Unwrap() error { return ErrorInvalidBundle }

type Input struct {
	Bundle *bytes.Buffer
}

type Output struct {
	ProjectID    int32
	Name         string
	ImageTag     string
	TestImageTag string
	// MissingImageTags is what the venue still has to build. Empty means the
	// project can judge as soon as it is attached to a contest.
	MissingImageTags []string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// ImageCatalogue answers which image tags this installation already has.
//
// Images belong to another bounded context, so this is a port: the binding
// lives in the composition root and the project domain never imports the
// evaluation environment.
type ImageCatalogue interface {
	ExistingTags(tags []string) (map[string]bool, error)
}

// ProjectNamer resolves a collision with an existing project's name.
type ProjectNamer interface {
	// UniqueName returns wanted, or wanted plus a suffix if a project already
	// holds it.
	UniqueName(wanted string) (string, error)
}

// AssetUploader and GradingFilesUploader are the two usecases that write to
// disk. Named as ports so a test can make either fail and watch the cleanup.
type AssetUploader interface {
	Upload(projectID int32, fileName string, file *bytes.Buffer) (string, error)
	// SetEnabled applies the flag after the upload, because uploading is the
	// only way to create an asset and it always starts disabled — the same
	// two steps the admin screens take.
	SetEnabled(assetID string, enabled bool) error
}

type GradingFilesUploader interface {
	Upload(projectID int32, zipped *bytes.Buffer) error
}

type usecase struct {
	zip           sharedservice.ZipService
	parsers       domainservice.ResultParserValidator
	images        ImageCatalogue
	namer         ProjectNamer
	createProject createproject.Usecase
	deleteProject deleteproject.Usecase
	createItem    creategradingitem.Usecase
	updateItem    updategradingitem.Usecase
	createCase    createtestcase.Usecase
	updateCase    updatetestcase.Usecase
	createLC      createlifecycle.Usecase
	assets        AssetUploader
	gradingFiles  GradingFilesUploader
}

func NewUsecase(
	zip sharedservice.ZipService,
	parsers domainservice.ResultParserValidator,
	images ImageCatalogue,
	namer ProjectNamer,
	createProject createproject.Usecase,
	deleteProject deleteproject.Usecase,
	createItem creategradingitem.Usecase,
	updateItem updategradingitem.Usecase,
	createCase createtestcase.Usecase,
	updateCase updatetestcase.Usecase,
	createLC createlifecycle.Usecase,
	assets AssetUploader,
	gradingFiles GradingFilesUploader,
) Usecase {
	return &usecase{
		zip: zip, parsers: parsers, images: images, namer: namer,
		createProject: createProject, deleteProject: deleteProject,
		createItem: createItem, updateItem: updateItem,
		createCase: createCase, updateCase: updateCase, createLC: createLC,
		assets: assets, gradingFiles: gradingFiles,
	}
}

// Execute imports a bundle in two halves: everything that can be checked is
// checked before anything is written, and only then does it create, following
// the same path a person clicking through the admin screens would take.
//
// The split is what makes the absence of a transaction tolerable. A project
// spans eight tables and two directories on disk, so no transaction could ever
// have covered all of it; what is achievable is that a REJECTED bundle never
// touches anything, leaving compensation to handle infrastructure faults —
// a database that goes away, a disk that fills — rather than user error.
func (u *usecase) Execute(input *Input) (*Output, error) {
	contents, definition, err := u.verify(input)
	if err != nil {
		return nil, err
	}

	return u.create(contents, definition)
}

// bundleContents is the archive, already read and checked.
type bundleContents struct {
	files map[string][]byte
}

func (u *usecase) verify(input *Input) (*bundleContents, *bundle.Project, error) {
	if input.Bundle == nil || input.Bundle.Len() == 0 {
		return nil, nil, &InvalidBundleError{Details: []string{"沒有收到檔案"}}
	}

	// The same hardened reader every other upload goes through: a bundle is
	// an untrusted archive from outside, and this one additionally carries
	// commands that will run on the judge host. ReadZipToMap bounds the total
	// size and entry count so a zip bomb cannot be expanded into memory here.
	files, err := u.zip.ReadZipToMap(input.Bundle)
	if err != nil {
		return nil, nil, &InvalidBundleError{Details: []string{"上傳的檔案不是有效或過大的 zip"}}
	}
	contents := &bundleContents{files: files}

	definitionJSON, ok := contents.files[bundle.ProjectFile]
	if !ok {
		return nil, nil, &InvalidBundleError{
			Details: []string{fmt.Sprintf("匯出檔裡沒有 %s", bundle.ProjectFile)},
		}
	}

	definition, err := bundle.Unmarshal(definitionJSON)
	if err != nil {
		return nil, nil, &InvalidBundleError{Details: []string{err.Error()}}
	}

	details := make([]string, 0)
	for _, p := range definition.Validate(u.parsers) {
		details = append(details, p.Error())
	}

	// Checked here rather than at upload time so a bundle missing a payload is
	// rejected before the first write, alongside every other fault, instead of
	// failing half way through and leaving a project to clean up.
	for i, a := range definition.Assets {
		if _, ok := contents.files[a.File]; !ok {
			details = append(details, fmt.Sprintf("assets[%d].file: 匯出檔裡找不到 %q", i, a.File))
		}
	}
	if definition.HasGradingFiles {
		gradingFiles, ok := contents.files[bundle.GradingFilesFile]
		if !ok {
			details = append(details, fmt.Sprintf("hasGradingFiles 為 true，但匯出檔裡沒有 %s", bundle.GradingFilesFile))
		} else if err := u.zip.CheckZip(bytes.NewBuffer(gradingFiles)); err != nil {
			details = append(details, fmt.Sprintf("%s 不是有效的 zip", bundle.GradingFilesFile))
		}
	}

	// Checked here, not after creating the project, because create_project
	// refuses a tag this installation does not have — as it should, and as it
	// would for a person filling in the same form. So a bundle naming an
	// unknown image cannot be imported at all, and saying so up front beats
	// failing half way and cleaning up. It also reads better for the operator:
	// "build these two images first" rather than "here is a project that
	// cannot judge".
	missing, err := u.missingImages(definition)
	if err != nil {
		return nil, nil, ErrorInternalServer
	}
	for _, tag := range missing {
		details = append(details, fmt.Sprintf("映像 %q 不存在於本系統，請先在「評分環境管理」建立並建置後再匯入", tag))
	}

	if len(details) > 0 {
		return nil, nil, &InvalidBundleError{Details: details}
	}

	return contents, definition, nil
}

// created records what has been written so far, so a failure can undo it.
//
// Only the project id, deliberately. Deleting a project leaves its grading
// items, testcases, lifecycles and assets behind — nothing in this system
// cascades — so an import that tidied up after itself more thoroughly would
// leave the library in a state the admin screens' own delete never produces.
// Two meanings of "this project is gone" is worse than one imperfect meaning,
// and the orphans are unreachable either way: every query reaches them through
// the project, which is soft-deleted.
type created struct {
	projectID int32
}

func (u *usecase) create(contents *bundleContents, definition *bundle.Project) (*Output, error) {
	name, err := u.namer.UniqueName(definition.Name)
	if err != nil {
		return nil, ErrorInternalServer
	}

	fields := make([]*project.DynamicFieldConfig, 0, len(definition.DynamicFields))
	for _, f := range definition.DynamicFields {
		fields = append(fields, &project.DynamicFieldConfig{
			Type:       f.Type,
			Name:       f.Name,
			Label:      f.Label,
			DefaultVal: f.DefaultValue,
			IsRequired: f.IsRequired,
		})
	}

	projectOut, err := u.createProject.Execute(&createproject.Input{
		Name:                name,
		Description:         definition.Description,
		ImageTag:            definition.ImageTag,
		TestImageTag:        definition.TestImageTag,
		Env:                 definition.Env,
		TestEnv:             definition.TestEnv,
		ResultParser:        definition.ResultParser,
		DynamicFieldConfigs: fields,
	})
	if err != nil {
		return nil, u.wrap(err)
	}

	done := &created{projectID: projectOut.ID}

	out, err := u.createChildren(contents, definition, done)
	if err != nil {
		return nil, u.rollback(done, err)
	}

	out.Name = name
	return out, nil
}

func (u *usecase) createChildren(contents *bundleContents, definition *bundle.Project, done *created) (*Output, error) {
	for _, l := range definition.Lifecycles {
		if _, err := u.createLC.Execute(&createlifecycle.Input{
			ProjectID:            done.projectID,
			Type:                 l.Type,
			Name:                 l.Name,
			Description:          l.Description,
			Command:              l.Command,
			ExecutionEnvironment: l.RunsOn,
		}); err != nil {
			return nil, err
		}
	}

	// Grading items are created in bundle order and take their position from
	// it: create_grading_item assigns the next order itself, which is why the
	// bundle carries no order field to disagree with.
	itemIDs := make([]int32, 0, len(definition.GradingItems))
	for _, g := range definition.GradingItems {
		itemOut, err := u.createItem.Execute(&creategradingitem.Input{
			ProjectID:       done.projectID,
			Name:            g.Name,
			Description:     g.Description,
			Score:           g.Score,
			ScoringStrategy: g.ScoringStrategy,
			OutputFormat:    g.OutputFormat,
			ResultParser:    g.ResultParser,
			Label:           g.Label,
		})
		if err != nil {
			return nil, err
		}
		itemIDs = append(itemIDs, itemOut.ID)
	}

	// Testcases first, dependencies after: a dependency may point at a
	// testcase declared later in the bundle, so no single pass could resolve
	// them all.
	idByKey := map[string]int32{}
	for i, g := range definition.GradingItems {
		for _, t := range g.Testcases {
			caseOut, err := u.createCase.Execute(&createtestcase.Input{
				GradingItemID: itemIDs[i],
				Name:          t.Name,
				Command:       t.Command,
				Score:         t.Score,
				IsSample:      t.IsSample,
			})
			if err != nil {
				return nil, err
			}
			idByKey[t.Key] = caseOut.ID
		}
	}

	for i, g := range definition.GradingItems {
		if len(g.DependsOnTestcases) > 0 {
			depends := resolve(g.DependsOnTestcases, idByKey)
			if _, err := u.updateItem.Execute(&updategradingitem.Input{
				ID:              itemIDs[i],
				DependTestcases: &depends,
			}); err != nil {
				return nil, err
			}
		}

		for _, t := range g.Testcases {
			if len(t.DependsOnTestcases) == 0 {
				continue
			}
			depends := resolve(t.DependsOnTestcases, idByKey)
			if _, err := u.updateCase.Execute(&updatetestcase.Input{
				ID:      idByKey[t.Key],
				Depends: &depends,
			}); err != nil {
				return nil, err
			}
		}
	}

	for _, a := range definition.Assets {
		id, err := u.assets.Upload(done.projectID, a.Name, bytes.NewBuffer(contents.files[a.File]))
		if err != nil {
			return nil, err
		}

		if a.IsEnabled {
			if err := u.assets.SetEnabled(id, true); err != nil {
				return nil, err
			}
		}
	}

	if definition.HasGradingFiles {
		if err := u.gradingFiles.Upload(done.projectID, bytes.NewBuffer(contents.files[bundle.GradingFilesFile])); err != nil {
			return nil, err
		}
	}

	return &Output{
		ProjectID:    done.projectID,
		ImageTag:     definition.ImageTag,
		TestImageTag: definition.TestImageTag,
		// Always empty on success: an import with a missing image is rejected
		// before it starts. The field stays in the response because the
		// composite bundle, which carries images and creates them first, has a
		// real use for it.
		MissingImageTags: []string{},
	}, nil
}

func (u *usecase) missingImages(definition *bundle.Project) ([]string, error) {
	wanted := []string{definition.ImageTag}
	if definition.TestImageTag != definition.ImageTag {
		wanted = append(wanted, definition.TestImageTag)
	}

	existing, err := u.images.ExistingTags(wanted)
	if err != nil {
		return nil, err
	}

	missing := make([]string, 0, len(wanted))
	for _, tag := range wanted {
		if !existing[tag] {
			missing = append(missing, tag)
		}
	}
	return missing, nil
}

// rollback removes what this import created, and reports what it could not.
//
// Best effort by necessity — the same absence of a transaction that makes it
// needed makes it fallible. What it must never do is hide the original
// failure: an operator told only "cleanup failed" cannot tell whether their
// bundle was wrong or the disk was full, and an operator told nothing at all
// is left with a half-built problem in the library and no idea it is there.
func (u *usecase) rollback(done *created, cause error) error {
	if _, err := u.deleteProject.Execute(&deleteproject.Input{ID: done.projectID}); err != nil {
		return fmt.Errorf("%w（另外，清理失敗：題目 #%d 仍留在系統中，請手動刪除）", u.wrap(cause), done.projectID)
	}
	return u.wrap(cause)
}

// wrap keeps a bundle-level fault reportable while collapsing everything else
// into an internal error: a failure from a create usecase at this point is a
// fault in this deployment, not something the operator's file can fix.
func (u *usecase) wrap(err error) error {
	var invalid *InvalidBundleError
	if errors.As(err, &invalid) {
		return err
	}
	return fmt.Errorf("%w: %s", ErrorInternalServer, strings.TrimSpace(err.Error()))
}

func resolve(keys []string, idByKey map[string]int32) []int32 {
	ids := make([]int32, 0, len(keys))
	for _, k := range keys {
		if id, ok := idByKey[k]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}
