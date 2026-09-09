package deleteimage

import (
	"errors"
	cancelbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/cancel_build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/image_service"
	"log"
)

var (
	ErrorTagIsRequired  = errors.New("tag is required")
	ErrorImageNotFound  = errors.New("image not found")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	Tag string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	imageRepository repository.ImageRepository
	imageService    imageservice.ImageService

	cancelBuildUsecase cancelbuild.Usecase
}

func NewUsecase(
	imageRepository repository.ImageRepository,
	imageService imageservice.ImageService,

	cancelBuildUsecase cancelbuild.Usecase,
) Usecase {
	return &usecase{
		imageRepository: imageRepository,
		imageService:    imageService,

		cancelBuildUsecase: cancelBuildUsecase,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	image, err := u.imageRepository.FindByTag(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if image == nil {
		return nil, ErrorImageNotFound
	}

	// Files FIRST, row second: while the files are being removed the
	// row still exists, so a concurrent create of the same tag is rejected by
	// the unique tag index (CreateExclusive) and can never write into the
	// directory being deleted. The old order (row, then files) opened a
	// window where a create could win the freed tag, write its files, and
	// have them wiped by this cleanup. Trade-off of the new order: if the
	// row delete below fails, the files are already gone but the row remains
	// — the user retries, and the second CleanupProject is a no-op RemoveAll
	// (a later create also wipes the dir itself: SetupProject begins with
	// CleanupProject).
	if err := u.imageService.CleanupProject(image); err != nil {
		// The files could not be removed, so the delete did NOT happen —
		// swallowing this used to answer success while leaving everything in
		// place.
		log.Printf("failed to cleanup project: %v", err)
		return nil, ErrorInternalServer
	}

	// cancelBuild BEFORE the row delete: Delete renames the soft-deleted
	// row's tag (freeing it on MySQL, whose unique index has no partial
	// WHERE), so a by-tag lookup after Delete would no longer resolve this
	// row. While the row is live the scoped lookup hits directly. Best-effort
	// log-only, exactly as it was after the delete.
	if _, err := u.cancelBuildUsecase.Execute(&cancelbuild.Input{Tag: input.Tag}); err != nil {
		log.Printf("failed to cancel build: %v", err)
	}

	if err := u.imageRepository.Delete(image); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
