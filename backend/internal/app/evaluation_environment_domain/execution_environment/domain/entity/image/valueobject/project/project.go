package imageproject

import (
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"path"
)

type Project struct {
	storagePath string
	tag         imagetag.Tag
}

func New(storagePath string, tag imagetag.Tag) (Project, error) {
	if err := tag.Validate(); err != nil {
		return Project{}, err
	}

	return Project{
		storagePath: path.Join(storagePath, "images", tag.String()),
		tag:         tag,
	}, nil
}

func (p Project) StoragePath() string {
	return p.storagePath
}

func (p Project) Tag() string {
	return p.tag.String()
}
