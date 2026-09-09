package imagequery

import (
	"judge-backend/internal/app/shared/model"
	imagequery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query/models/imageinfo"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(
	db *gorm.DB,
) imagequery.ImageQueryService {
	return &Service{
		db: db,
	}
}

func (s *Service) GetImages() ([]imageinfo.ImageInfo, error) {
	images := []model.Image{}
	if err := s.db.Find(&images).Error; err != nil {
		return nil, err
	}

	imageInfos := make([]imageinfo.ImageInfo, 0, len(images))
	for _, img := range images {
		imageInfos = append(imageInfos, imageinfo.ImageInfo{
			Tag:         img.Tag,
			Description: img.Description,
			IsReady:     img.IsReady,
			CreatedAt:   img.CreatedAt,
			UpdatedAt:   img.UpdatedAt,
		})
	}

	return imageInfos, nil
}
