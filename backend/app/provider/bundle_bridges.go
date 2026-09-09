package provider

import (
	"errors"
	"path"

	contestbundle "judge-backend/internal/app/contest_domain/contest/application/bundle"
	exportcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/export_contest"
	importcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/import_contest"
	imagebundle "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/bundle"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/fingerprint"
	exportimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/export_image"
	importimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/import_image"
	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
	importbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/import_bundle"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
	projectbundle "judge-backend/internal/app/project_domain/project/application/bundle"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
	importproject "judge-backend/internal/app/project_domain/project/application/usecase/project/import_project"
	"judge-backend/internal/app/shared/model"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/config"
	"judge-backend/internal/projectpath"

	"bytes"

	"gorm.io/gorm"
)

// The composite bundle's bridges. They live here, in the composition root,
// because the composite bundle spans three bounded contexts and none of them
// may import another: each adapter below calls one domain's own usecases or
// reads shared tables, and the portability BC sees only its ports.

// bundleBridges returns every binding the portability BC needs.
func bundleBridges() []interface{} {
	return []interface{}{
		func(u exportcontest.Usecase) exportbundle.ContestExporter { return contestExporterBridge{u} },
		func(db *gorm.DB) exportbundle.ContestProjectLister { return contestProjectLister{db} },
		func(u exportproject.Usecase) exportbundle.ProjectExporter { return projectExporterBridge{u} },
		func(u exportimage.Usecase) exportbundle.ImageExporter { return imageExporterBridge{u} },
		func() previewbundle.PartReader { return partReader{} },
		func(db *gorm.DB, fs sharedservice.FileSystemService, cfg *config.Config) previewbundle.ImageInspector {
			return imageInspector{db: db, fs: fs, storagePath: path.Join(projectpath.Root, cfg.Server.StoragePath)}
		},
		func(db *gorm.DB) previewbundle.ProjectInspector { return projectInspector{db} },
		func(u importimage.Usecase) importbundle.ImageImporter { return imageImporterBridge{u} },
		func(u importproject.Usecase) importbundle.ProjectImporter { return projectImporterBridge{u} },
		func(u importcontest.Usecase) importbundle.ContestImporter { return contestImporterBridge{u} },
	}
}

// ── part readers ────────────────────────────────────────────────────────────

// partReader parses each part with the package that owns its format. The
// portability BC deliberately has no parser of its own: a second reader of
// project.json would drift from the first, and this is the one place allowed
// to import all three.
type partReader struct{}

func (partReader) ReadProject(definition []byte) (string, []string, error) {
	p, err := projectbundle.Unmarshal(definition)
	if err != nil {
		return "", nil, err
	}
	tags := []string{p.ImageTag}
	if p.TestImageTag != p.ImageTag {
		tags = append(tags, p.TestImageTag)
	}
	return p.Name, tags, nil
}

func (partReader) ReadImage(definition []byte) (string, string, string, error) {
	i, err := imagebundle.Unmarshal(definition)
	if err != nil {
		return "", "", "", err
	}
	return i.Tag, i.Mode, i.Fingerprint, nil
}

func (partReader) ReadContest(definition []byte) (string, []string, error) {
	c, err := contestbundle.Unmarshal(definition)
	if err != nil {
		return "", nil, err
	}
	refs := make([]string, 0, len(c.Projects))
	for _, p := range c.Projects {
		refs = append(refs, p.ProjectRef)
	}
	return c.Name, refs, nil
}

// ── preview inspectors ──────────────────────────────────────────────────────

type imageInspector struct {
	db          *gorm.DB
	fs          sharedservice.FileSystemService
	storagePath string
}

// Inspect mirrors the decision the image import itself makes (sameContent in
// import_image): only a source bundle whose build context is still on disk can
// be proven identical; an archive's digest never matches one built elsewhere,
// so an existing tag is a conflict there. Mirrored rather than shared because
// the import decides for one image mid-flight and this answers for a listing —
// but the RULE must stay the same, or the preview would promise an import that
// then refuses.
func (i imageInspector) Inspect(tag, mode, print string) (string, []string, error) {
	var count int64
	if err := i.db.Model(&model.Image{}).Where("tag = ?", tag).Count(&count).Error; err != nil {
		return "", nil, err
	}
	if count == 0 {
		return previewbundle.StatusNew, nil, nil
	}

	if mode == imagebundle.ModeSource {
		contextPath := path.Join(i.storagePath, "images", tag)
		if i.fs.Exists(contextPath) {
			existing, err := fingerprint.OfBuildContext(i.fs, contextPath)
			if err == nil && existing == print {
				return previewbundle.StatusIdentical, nil, nil
			}
		}
	}

	var usedBy []string
	if err := i.db.Model(&model.Project{}).
		Where("image_tag = ? OR test_image_tag = ?", tag, tag).
		Order("id ASC").
		Pluck("name", &usedBy).Error; err != nil {
		return "", nil, err
	}
	return previewbundle.StatusConflict, usedBy, nil
}

type projectInspector struct{ db *gorm.DB }

func (p projectInspector) NameExists(name string) (bool, error) {
	var count int64
	if err := p.db.Model(&model.Project{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p projectInspector) TagsPresent(tags []string) (map[string]bool, error) {
	found := make(map[string]bool, len(tags))
	if len(tags) == 0 {
		return found, nil
	}
	var rows []model.Image
	if err := p.db.Where("tag IN ?", tags).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		found[r.Tag] = true
	}
	return found, nil
}

// ── exporters ───────────────────────────────────────────────────────────────

type contestExporterBridge struct{ u exportcontest.Usecase }

func (b contestExporterBridge) Export(contestID int32) ([]byte, error) {
	out, err := b.u.Execute(&exportcontest.Input{ContestID: contestID})
	if errors.Is(err, exportcontest.ErrorContestNotFound) {
		return nil, exportbundle.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return out.Content, nil
}

type contestProjectLister struct{ db *gorm.DB }

// ProjectIDs walks the slots in the same order the contest export writes them
// (id ascending), so the composite's projects/ directory lines up with the
// contest.json it sits next to.
func (l contestProjectLister) ProjectIDs(contestID int32) ([]int32, error) {
	var ids []int32
	if err := l.db.Model(&model.ContestProject{}).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Pluck("project_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

type projectExporterBridge struct{ u exportproject.Usecase }

func (b projectExporterBridge) Export(projectID int32) ([]byte, error) {
	out, err := b.u.Execute(&exportproject.Input{ProjectID: projectID})
	if err != nil {
		return nil, err
	}
	return out.Content, nil
}

type imageExporterBridge struct{ u exportimage.Usecase }

func (b imageExporterBridge) Export(tag, mode string) ([]byte, error) {
	out, err := b.u.Execute(&exportimage.Input{Tag: tag, Mode: mode})
	if errors.Is(err, exportimage.ErrorNotBuilt) {
		return nil, exportbundle.ErrImageNotBuilt
	}
	if err != nil {
		return nil, err
	}
	return out.Content, nil
}

// ── importers ───────────────────────────────────────────────────────────────

type imageImporterBridge struct{ u importimage.Usecase }

func (b imageImporterBridge) Import(partZip []byte, force bool) (bool, error) {
	out, err := b.u.Execute(&importimage.Input{Bundle: bytes.NewBuffer(partZip), Force: force})
	if err != nil {
		return false, err
	}
	return out.BuildQueued, nil
}

type projectImporterBridge struct{ u importproject.Usecase }

func (b projectImporterBridge) Import(partZip []byte) (int32, string, error) {
	out, err := b.u.Execute(&importproject.Input{Bundle: bytes.NewBuffer(partZip)})
	if err != nil {
		return 0, "", err
	}
	return out.ProjectID, out.Name, nil
}

type contestImporterBridge struct{ u importcontest.Usecase }

func (b contestImporterBridge) Import(partZip []byte) (int32, int32, []string, error) {
	out, err := b.u.Execute(&importcontest.Input{Bundle: bytes.NewBuffer(partZip)})
	if err != nil {
		return 0, 0, nil, err
	}
	return out.ContestID, out.CreatedProjects, out.UnresolvedProjects, nil
}
