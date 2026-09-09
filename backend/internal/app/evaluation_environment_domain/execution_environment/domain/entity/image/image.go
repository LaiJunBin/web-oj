package image

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	imagedescription "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/description"
	imagemountpath "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/mountpath"
	imageproject "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/project"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

type Image struct {
	tag         imagetag.Tag
	description imagedescription.Description
	project     imageproject.Project
	mountPath   imagemountpath.MountPath
	isReady     bool
	sharedtypes.TimeStamps
}

type Config struct {
	Tag         string
	Description string
	MountPath   string
	StoragePath string
}

func New(config *Config) (*Image, error) {
	tag, err := imagetag.New(config.Tag)
	if err != nil {
		return nil, err
	}

	project, err := imageproject.New(config.StoragePath, tag)
	if err != nil {
		return nil, err
	}

	mountPath, err := imagemountpath.New(config.MountPath)
	if err != nil {
		return nil, err
	}

	description := imagedescription.New(config.Description)
	return &Image{
		tag:         tag,
		description: description,
		project:     project,
		mountPath:   mountPath,
		isReady:     false,
	}, nil
}

func MustNew(config *Config) *Image {
	image, err := New(config)
	if err != nil {
		panic(err)
	}

	return image
}

func (i *Image) Tag() imagetag.Tag {
	return i.tag
}

func (i *Image) SetTag(tag string) error {
	imageTag, err := imagetag.New(tag)
	if err != nil {
		return err
	}

	i.tag = imageTag
	return nil
}

func (i *Image) Description() string {
	return i.description.String()
}

func (i *Image) SetDescription(description string) {
	i.description = imagedescription.New(description)
}

func (i *Image) Project() imageproject.Project {
	return i.project
}

func (i *Image) MountPath() imagemountpath.MountPath {
	return i.mountPath
}

func (i *Image) SetMountPath(mountPath string) error {
	imageMountPath, err := imagemountpath.New(mountPath)
	if err != nil {
		return err
	}

	i.mountPath = imageMountPath
	return nil
}

func (i *Image) IsReady() bool {
	return i.isReady
}

func (i *Image) MarkAsReady() {
	i.isReady = true
}

func (i *Image) MarkAsNotReady() {
	i.isReady = false
}
