package service

import (
	"bytes"
	"fmt"
	"time"

	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	updateasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/update_asset"
	uploadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/upload_asset"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
	importproject "judge-backend/internal/app/project_domain/project/application/usecase/project/import_project"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

// The small adapters that let the import usecase talk to things outside its
// own package without importing them: the image catalogue lives in another
// bounded context, and the two uploaders are usecases whose signatures the
// importer should not depend on directly.

type imageCatalogue struct{ db *gorm.DB }

// NewImageCatalogue reports which image tags this installation holds.
//
// It reads the images table directly rather than calling the evaluation
// environment's usecases: this is a membership question, and routing it
// through another bounded context's application layer would couple the two for
// no gain. The binding is still declared as a port so the dependency direction
// stays visible in the composition root.
func NewImageCatalogue(db *gorm.DB) importproject.ImageCatalogue {
	return &imageCatalogue{db: db}
}

func (c *imageCatalogue) ExistingTags(tags []string) (map[string]bool, error) {
	found := make(map[string]bool, len(tags))
	if len(tags) == 0 {
		return found, nil
	}

	var rows []model.Image
	if err := c.db.Where("tag IN ?", tags).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		found[r.Tag] = true
	}
	return found, nil
}

type projectNamer struct{ db *gorm.DB }

func NewProjectNamer(db *gorm.DB) importproject.ProjectNamer {
	return &projectNamer{db: db}
}

// UniqueName appends the import time when the name is already taken.
//
// Project names are not unique in the database, so this is a courtesy rather
// than a constraint — but two identical rows in the admin list help nobody
// choose between them. The suffix says when, which is the one fact that
// distinguishes two imports of the same problem.
func (n *projectNamer) UniqueName(wanted string) (string, error) {
	taken, err := n.exists(wanted)
	if err != nil {
		return "", err
	}
	if !taken {
		return wanted, nil
	}

	candidate := fmt.Sprintf("%s (匯入 %s)", wanted, time.Now().Format("2006-01-02 15:04"))
	taken, err = n.exists(candidate)
	if err != nil {
		return "", err
	}
	if !taken {
		return candidate, nil
	}

	// Two imports of the same problem within one minute. Rare, but the second
	// one still has to land somewhere.
	for i := 2; ; i++ {
		numbered := fmt.Sprintf("%s-%d", candidate, i)
		taken, err := n.exists(numbered)
		if err != nil {
			return "", err
		}
		if !taken {
			return numbered, nil
		}
	}
}

func (n *projectNamer) exists(name string) (bool, error) {
	var count int64
	if err := n.db.Model(&model.Project{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type assetUploader struct {
	upload uploadasset.Usecase
	update updateasset.Usecase
}

func NewAssetUploader(upload uploadasset.Usecase, update updateasset.Usecase) importproject.AssetUploader {
	return &assetUploader{upload: upload, update: update}
}

func (a *assetUploader) Upload(projectID int32, fileName string, file *bytes.Buffer) (string, error) {
	out, err := a.upload.Execute(&uploadasset.Input{
		ProjectID: projectID,
		FileName:  fileName,
		File:      file,
	})
	if err != nil {
		return "", err
	}
	return out.ID, nil
}

func (a *assetUploader) SetEnabled(assetID string, enabled bool) error {
	_, err := a.update.Execute(&updateasset.Input{ID: assetID, IsEnabled: &enabled})
	return err
}

type gradingFilesUploader struct {
	service gradingproject.GradingProjectService
}

func NewGradingFilesUploader(s gradingproject.GradingProjectService) importproject.GradingFilesUploader {
	return &gradingFilesUploader{service: s}
}

func (g *gradingFilesUploader) Upload(projectID int32, zipped *bytes.Buffer) error {
	return g.service.UploadGradingProject(projectID, zipped)
}

type systemClock struct{}

// NewExportClock supplies the export's provenance timestamp. Injected so the
// one non-deterministic input to an export can be pinned in a test.
func NewExportClock() exportproject.Clock { return systemClock{} }

func (systemClock) NowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }
