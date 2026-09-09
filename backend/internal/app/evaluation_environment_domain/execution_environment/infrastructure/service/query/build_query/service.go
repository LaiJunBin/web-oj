package buildquery

import (
	"errors"
	"judge-backend/internal/app/shared/model"
	buildquery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query/models/buildinfo"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"

	"gorm.io/gorm"
)

type Service struct {
	repo   repository.BuildRepository
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewService(
	repo repository.BuildRepository,
	db *gorm.DB,
	logger sharedtypes.Logger,
) buildquery.BuildQueryService {
	return &Service{
		repo:   repo,
		db:     db,
		logger: logger,
	}
}

func (s *Service) GetBuilds(tag string) ([]buildinfo.BuildInfo, error) {
	image := &model.Image{}
	if err := s.db.First(image, "tag = ?", tag).Error; err != nil {
		// Asking for an unknown tag is an ordinary client mistake, not a
		// server fault — only a real database error deserves Errorf.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("[BuildQueryService] image %s not found: %v", tag, err)
		} else {
			s.logger.Errorf("[BuildQueryService] failed to load image for tag %s: %v", tag, err)
		}
		return nil, err
	}

	builds := []model.Build{}
	if err := s.db.Find(&builds, "image_id = ?", image.ID).Error; err != nil {
		s.logger.Errorf("[BuildQueryService] failed to load builds for image %s: %v", tag, err)
		return nil, err
	}

	buildInfos := make([]buildinfo.BuildInfo, 0, len(builds))
	for _, b := range builds {
		status, err := buildstatus.New(b.Status)
		if err != nil {
			s.logger.Errorf("[BuildQueryService] build %d of image %s has an invalid status %q: %v", b.Batch, tag, b.Status, err)
			return nil, err
		}
		buildInfos = append(buildInfos, buildinfo.BuildInfo{
			Batch:  int32(b.Batch),
			Status: status.String(),
		})
	}

	currentBuild := s.repo.FindCurrentBuildByTag(tag)
	if currentBuild != nil {
		buildInfos = append(buildInfos, buildinfo.BuildInfo{
			Batch:  int32(currentBuild.Batch().Value()),
			Status: currentBuild.Status().String(),
		})
	}

	return buildInfos, nil
}
