package imagequery

import "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query/models/imageinfo"

type ImageQueryService interface {
	GetImages() ([]imageinfo.ImageInfo, error)
}
