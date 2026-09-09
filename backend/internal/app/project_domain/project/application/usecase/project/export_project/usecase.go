package exportproject

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"judge-backend/internal/app/project_domain/project/application/bundle"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/zipfile"
)

var (
	ErrorInvalidProjectID = errors.New("invalid project id, must be greater than 0")
	ErrorProjectNotFound  = errors.New("project not found")
	ErrorInternalServer   = errors.New("internal server error")
)

type Input struct {
	ProjectID int32
}

type Output struct {
	Filename string
	Content  []byte
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// Clock is injected so the provenance timestamp is controllable in tests. It
// is the only non-deterministic input to an export, and it deliberately lives
// outside project.json for that reason.
type Clock interface {
	NowRFC3339() string
}

type usecase struct {
	query        query.ProjectExportQuery
	assets       query.AssetQuery
	gradingFiles gradingproject.GradingProjectService
	zip          sharedservice.ZipService
	clock        Clock
	logger       sharedtypes.Logger
}

func NewUsecase(
	q query.ProjectExportQuery,
	assets query.AssetQuery,
	gradingFiles gradingproject.GradingProjectService,
	zip sharedservice.ZipService,
	clock Clock,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{query: q, assets: assets, gradingFiles: gradingFiles, zip: zip, clock: clock, logger: logger}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if input.ProjectID <= 0 {
		return nil, ErrorInvalidProjectID
	}

	loaded, err := u.query.LoadProjectForExport(input.ProjectID)
	if errors.Is(err, query.ErrorProjectNotFound) {
		return nil, ErrorProjectNotFound
	}
	if err != nil {
		return nil, u.fail("載入題目 %d 失敗", input.ProjectID, err)
	}

	files := make([]zipfile.File, 0, len(loaded.Assets)+3)

	definition, assetFiles, err := u.buildDefinition(loaded)
	if err != nil {
		return nil, err
	}
	files = append(files, assetFiles...)

	// Grading files are part of the problem, not an attachment: they are
	// copied into the test container at judge time, so a bundle without them
	// grades differently. Three outcomes, and only one of them is quiet:
	// none uploaded is a legitimate state and the bundle simply says so;
	// but files that exist and cannot be read must fail the export loudly —
	// silently producing a bundle without them would hand the importer a
	// problem that grades differently from the one being exported.
	gradingFiles, err := u.gradingFiles.DownloadGradingProject(loaded.ID)
	switch {
	case errors.Is(err, gradingproject.ErrorGradingProjectNotFound):
		// No grading files for this project; nothing to include.
	case err != nil:
		return nil, u.fail("讀取題目 %d 的評測檔失敗", loaded.ID, err)
	case gradingFiles != nil && len(gradingFiles.Content) > 0:
		definition.HasGradingFiles = true
		files = append(files, zipfile.File{
			Path: bundle.GradingFilesFile,
			Body: string(gradingFiles.Content),
		})
	}

	definitionJSON, err := bundle.Marshal(definition)
	if err != nil {
		return nil, u.fail("序列化題目 %d 的定義失敗", loaded.ID, err)
	}

	meta, err := bundle.MarshalMeta(&bundle.Meta{
		ExportedAt:      u.clock.NowRFC3339(),
		SourceProjectID: loaded.ID,
	})
	if err != nil {
		return nil, u.fail("序列化題目 %d 的 meta 失敗", loaded.ID, err)
	}

	files = append(files,
		zipfile.File{Path: bundle.ProjectFile, Body: string(definitionJSON)},
		zipfile.File{Path: bundle.MetaFile, Body: string(meta)},
	)

	archive, err := u.zip.ZipFromFiles(&files)
	if err != nil {
		return nil, u.fail("打包題目 %d 失敗", loaded.ID, err)
	}

	return &Output{
		Filename: filenameFor(loaded.Name, loaded.ID),
		Content:  archive.Bytes(),
	}, nil
}

// buildDefinition converts the loaded project into the bundle schema and
// collects the asset payloads that go alongside it.
func (u *usecase) buildDefinition(loaded *dto.ProjectExport) (*bundle.Project, []zipfile.File, error) {
	out := &bundle.Project{
		FormatVersion: bundle.FormatVersion,
		Kind:          bundle.KindProject,
		Name:          loaded.Name,
		Description:   loaded.Description,
		ImageTag:      loaded.ImageTag,
		TestImageTag:  loaded.TestImageTag,
		Env:           loaded.Env,
		TestEnv:       loaded.TestEnv,
		ResultParser:  loaded.ResultParser,
		DynamicFields: make([]bundle.DynamicField, 0, len(loaded.DynamicFields)),
		Lifecycles:    make([]bundle.Lifecycle, 0, len(loaded.Lifecycles)),
		GradingItems:  make([]bundle.GradingItem, 0, len(loaded.GradingItems)),
		Assets:        make([]bundle.Asset, 0, len(loaded.Assets)),
	}

	for _, f := range loaded.DynamicFields {
		out.DynamicFields = append(out.DynamicFields, bundle.DynamicField{
			Type:         f.Type,
			Name:         f.Name,
			Label:        f.Label,
			DefaultValue: f.DefaultValue,
			IsRequired:   f.IsRequired,
		})
	}

	for _, l := range loaded.Lifecycles {
		out.Lifecycles = append(out.Lifecycles, bundle.Lifecycle{
			Name:        l.Name,
			Description: l.Description,
			Type:        l.Type,
			RunsOn:      l.RunsOn,
			Command:     l.Command,
		})
	}

	// Dependencies are stored as database ids, which mean nothing on another
	// installation. Every testcase gets a bundle key first, so the second
	// pass can translate ids into keys — including forward references.
	keyByID := map[int32]string{}
	for i, g := range loaded.GradingItems {
		for j, t := range g.Testcases {
			keyByID[t.ID] = fmt.Sprintf("g%dt%d", i+1, j+1)
		}
	}

	for _, g := range loaded.GradingItems {
		item := bundle.GradingItem{
			Label:              g.Label,
			Name:               g.Name,
			Description:        g.Description,
			Score:              g.Score,
			ScoringStrategy:    g.ScoringStrategy,
			OutputFormat:       g.OutputFormat,
			ResultParser:       g.ResultParser,
			DependsOnTestcases: keysFor(g.DependTestcaseIDs, keyByID),
			Testcases:          make([]bundle.Testcase, 0, len(g.Testcases)),
		}
		for _, t := range g.Testcases {
			item.Testcases = append(item.Testcases, bundle.Testcase{
				Key:                keyByID[t.ID],
				Name:               t.Name,
				Command:            t.Command,
				Score:              t.Score,
				IsSample:           t.IsSample,
				DependsOnTestcases: keysFor(t.DependTestcaseIDs, keyByID),
			})
		}
		out.GradingItems = append(out.GradingItems, item)
	}

	assetFiles := make([]zipfile.File, 0, len(loaded.Assets))
	for i, a := range loaded.Assets {
		content, err := u.assets.DownloadAsset(a.ID)
		if err != nil {
			// An asset row whose file is gone would otherwise produce a
			// bundle that imports into a project with a broken download.
			u.logger.Errorf("[ExportProject] asset %s (%q) of project %d could not be read: %v", a.ID, a.Name, loaded.ID, err)
			return nil, nil, fmt.Errorf("%w: 附件 %q 的檔案讀取失敗", ErrorInternalServer, a.Name)
		}

		// Named by position, not by the asset's display name: a display name
		// may repeat, may contain a path separator, and may be in a script the
		// receiving filesystem cannot represent.
		inZip := path.Join(bundle.AssetsDir, fmt.Sprintf("%d%s", i+1, path.Ext(a.Name)))

		out.Assets = append(out.Assets, bundle.Asset{
			Name:      a.Name,
			IsEnabled: a.IsEnabled,
			File:      inZip,
		})
		assetFiles = append(assetFiles, zipfile.File{Path: inZip, Body: string(content.Content)})
	}

	return out, assetFiles, nil
}

// keysFor translates stored dependency ids into bundle keys, dropping any that
// point outside this project. A dangling id cannot be expressed in the bundle
// and would fail its own validation on the way back in.
func keysFor(ids []int32, keyByID map[int32]string) []string {
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		if key, ok := keyByID[id]; ok {
			keys = append(keys, key)
		}
	}
	return keys
}

// filenameFor is what the browser saves. The project name is only a hint —
// it can be in any script and contain anything — so the id is what actually
// distinguishes two downloads.
func filenameFor(name string, id int32) string {
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		case r == ' ':
			return '-'
		default:
			return -1
		}
	}, name)
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "project"
	}
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	return fmt.Sprintf("%s-%d.zip", slug, id)
}

// fail records what actually went wrong and returns the opaque error.
//
// The response deliberately says only "internal server error": a 500 body is
// not the place for schema details. But collapsing the cause there must not
// also erase it — an export that fails with nothing in the log leaves whoever
// is on call with a screenshot of a generic message and no way forward. That
// is exactly how the first bug in this usecase was diagnosed, by adding a
// temporary print.
func (u *usecase) fail(format string, id int32, cause error) error {
	u.logger.Errorf("[ExportProject] "+format+": %v", id, cause)
	return ErrorInternalServer
}
