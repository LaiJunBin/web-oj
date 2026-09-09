package query

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"gorm.io/gorm"
)

type ProjectQuery struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewProjectQuery(
	db *gorm.DB,
	logger sharedtypes.Logger,
) query.ProjectQuery {
	return &ProjectQuery{
		db:     db,
		logger: logger,
	}
}

func (q ProjectQuery) ListProject() ([]dto.Project, error) {
	projects := []model.Project{}
	if err := q.db.Find(&projects).Error; err != nil {
		return nil, err
	}

	projectInfos := make([]dto.Project, 0, len(projects))
	for _, project := range projects {
		projectInfos = append(projectInfos, dto.Project{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			Image:       project.ImageTag,
			TestImage:   project.TestImageTag,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		})
	}

	return projectInfos, nil
}

func (q ProjectQuery) GetProjectByID(id int32) (*dto.ProjectDetail, error) {
	modelProject := model.Project{}
	if err := q.db.
		Preload("Image").
		Preload("TestImage").
		Preload("Assets").
		Preload("DynamicFields").
		Preload("Lifecycles").
		First(&modelProject, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var dtoImage *dto.Image
	if modelProject.Image != nil {
		dtoImage = &dto.Image{
			Tag:         modelProject.Image.Tag,
			Description: modelProject.Image.Description,
			IsReady:     modelProject.Image.IsReady,
		}
	}

	var dtoTestImage *dto.Image
	if modelProject.TestImage != nil {
		dtoTestImage = &dto.Image{
			Tag:         modelProject.TestImage.Tag,
			Description: modelProject.TestImage.Description,
			IsReady:     modelProject.TestImage.IsReady,
		}
	}

	dtoAssets := []*dto.Asset{}
	for _, asset := range modelProject.Assets {
		dtoAssets = append(dtoAssets, &dto.Asset{
			ID:           asset.ID,
			Name:         asset.Name,
			Size:         asset.Size,
			Enabled:      asset.IsEnabled,
			LastModified: asset.UpdatedAt,
		})
	}

	dtoDynamicFields := []*dto.DynamicField{}
	for _, dynamicField := range modelProject.DynamicFields {
		dtoDynamicFields = append(dtoDynamicFields, &dto.DynamicField{
			Type:     dynamicField.Type,
			Name:     dynamicField.Name,
			Label:    dynamicField.Label,
			Default:  dynamicField.DefaultValue,
			Required: dynamicField.IsRequired,
		})
	}

	dtoLifecycle := &dto.Lifecycle{}
	dtoTestLifecycle := &dto.Lifecycle{}
	for _, lc := range modelProject.Lifecycles {
		cmd := &dto.Command{
			ID:          lc.ID,
			Name:        lc.Name,
			Description: lc.Description,
			Text:        lc.Command,
		}

		var l *dto.Lifecycle
		switch {
		case lc.RunsOn == lifecycle.ExecutionEnvironmentProject.String():
			l = dtoLifecycle
		case lc.RunsOn == lifecycle.ExecutionEnvironmentTest.String():
			l = dtoTestLifecycle
		default:
			// A lifecycle row with an unrecognized runs_on is DATA the query
			// silently drops — say so where log collection can see it.
			q.logger.Warnf("[ProjectQuery] unknown lifecycle runs_on %q (lifecycle %d), skipping the command", lc.RunsOn, lc.ID)
			continue
		}

		switch {
		case lc.Type == lifecycle.BEFORE_ALL:
			l.BeforeAll = append(l.BeforeAll, cmd)
		case lc.Type == lifecycle.BEFORE_EACH:
			l.BeforeEach = append(l.BeforeEach, cmd)
		case lc.Type == lifecycle.AFTER_ALL:
			l.AfterAll = append(l.AfterAll, cmd)
		case lc.Type == lifecycle.AFTER_EACH:
			l.AfterEach = append(l.AfterEach, cmd)
		}
	}

	resultParser := modelProject.ResultParser
	if resultParser == "" {
		resultParser = resultparser.AutoParser.String()
	}

	return &dto.ProjectDetail{
		ID:            modelProject.ID,
		Name:          modelProject.Name,
		Description:   modelProject.Description,
		Image:         dtoImage,
		TestImage:     dtoTestImage,
		Env:           modelProject.Env,
		TestEnv:       modelProject.TestEnv,
		ResultParser:  resultParser,
		Assets:        dtoAssets,
		DynamicFields: dtoDynamicFields,
		Lifecycle:     *dtoLifecycle,
		TestLifecycle: *dtoTestLifecycle,
		CreatedAt:     modelProject.CreatedAt,
		UpdatedAt:     modelProject.UpdatedAt,
	}, nil
}
