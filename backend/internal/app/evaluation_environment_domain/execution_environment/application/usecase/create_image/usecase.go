package createimage

import (
	"bytes"
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	imageentity "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	imagemountpath "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/mountpath"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/image_service"
	"log"
)

var (
	ErrorInvalidInput         = errors.New("invalid input")
	ErrorTagIsRequired        = errors.New("tag is required")
	ErrorProjectZipIsRequired = errors.New("project zip file is required")
	ErrorInvalidTag           = errors.New("invalid tag")
	ErrorInvalidMountPath     = errors.New("invalid mount path")
	ErrorImageAlreadyExists   = errors.New("image already exists")
	ErrorFailedToSetup        = errors.New("failed to setup project")
	ErrorInternalServer       = errors.New("internal server error")
)

type Input struct {
	Tag            string
	MountPath      string
	Description    string
	ProjectZipFile *bytes.Buffer
	NoCache        bool
}

type Output struct {
	Tag string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	imageFactory    *factory.ImageFactory
	imageRepository repository.ImageRepository
	imageService    imageservice.ImageService
	buildService    buildservice.BuildService
}

func NewUsecase(
	imageFactory *factory.ImageFactory,
	imageRepository repository.ImageRepository,
	imageService imageservice.ImageService,
	buildService buildservice.BuildService,
) Usecase {
	return &usecase{
		imageFactory:    imageFactory,
		imageRepository: imageRepository,
		imageService:    imageService,
		buildService:    buildService,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	if input.ProjectZipFile == nil {
		return ErrorProjectZipIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	if image, err := u.imageRepository.FindByTag(input.Tag); err != nil {
		return nil, ErrorInternalServer
	} else if image != nil {
		return nil, ErrorImageAlreadyExists
	}

	image, err := u.imageFactory.CreateImage(input.Tag, input.MountPath, input.Description)
	if err != nil {
		switch {
		case errors.Is(err, imagetag.ErrInvalidTagFormat):
			return nil, ErrorInvalidTag
		case errors.Is(err, imagemountpath.ErrorInvalidMountPath):
			return nil, ErrorInvalidMountPath
		default:
			return nil, ErrorInternalServer
		}
	}

	// Two-phase creation, ownership first: (1) CreateExclusive lets the
	// database's unique tag index arbitrate concurrent creates — of N racing
	// requests exactly one owns the tag from here on, so only that one may
	// touch the shared project path. The old order (setup → save) let every
	// racer RemoveAll/Unzip the same directory before the constraint had
	// picked a winner. (2) The build-slot reservation, (3) the project
	// setup, and (4) the enqueue then all happen under that ownership, with
	// each failure compensating exactly what THIS request created — the
	// loser of the create never reaches any compensation, so a winner's row,
	// reservation, or files can never be deleted by a losing request.
	ownershipToken, err := u.imageRepository.CreateExclusive(image)
	if err != nil {
		if errors.Is(err, repository.ErrImageTagAlreadyExists) {
			return nil, ErrorImageAlreadyExists
		}
		return nil, ErrorInternalServer
	}

	reserved, err := u.buildService.ReserveBuild(image)
	if err != nil {
		// No reservation was acquired and no files were written — only the
		// image row this request just created needs removing, or a retry
		// would forever get ErrorImageAlreadyExists for an image that never
		// got a build.
		u.deleteOwnImage(image, ownershipToken)
		return nil, ErrorInternalServer
	}

	if err := u.imageService.SetupProject(image, input.ProjectZipFile); err != nil {
		u.compensate(image, reserved, ownershipToken)
		return nil, ErrorFailedToSetup
	}

	if err := u.buildService.EnqueueReservedBuild(reserved, image, &buildservice.BuildOptions{
		NoCache: input.NoCache,
	}); err != nil {
		// Never swallow the enqueue failure: reporting success without a
		// build job would leave the user waiting for an image that will
		// never become ready.
		log.Printf("failed to enqueue build image(%s) job: %v", image.Tag(), err)
		u.compensate(image, reserved, ownershipToken)
		return nil, ErrorInternalServer
	}

	output := &Output{
		Tag: image.Tag().String(),
	}
	return output, nil
}

// compensate unwinds a create whose setup or enqueue failed: the build-slot
// reservation (identity-guarded release), the image row this request
// inserted, and the project files this request wrote. The row is deleted via
// the ownership token from CreateExclusive, never by tag: the tag can be
// deleted and re-created by another request while this one is still in
// flight, and a by-tag delete would then land on the NEW owner's row.
//
// Row BEFORE files, and files only on a confirmed delete: the project
// PATH is derived from the tag, so it is shared with whoever owns the tag
// next. DeleteIfOwned's verdict is the fence — true means this request still
// owned the tag an instant ago, so the files at that path are its own to
// remove; false means the tag already changed hands (deleted and re-created
// while this create was in flight) and the path now holds the NEW owner's
// files, which must be left alone. The window this closes used to span the
// whole in-flight time of the create; what remains is the instants between
// the row delete and the RemoveAll — closing that too would need each attempt
// to stage its files in its own directory and promote them atomically on
// success.
func (u *usecase) compensate(image *imageentity.Image, reserved *build.Build, ownershipToken int) {
	u.buildService.ReleaseReservedBuild(reserved)

	owned, err := u.imageRepository.DeleteIfOwned(ownershipToken)
	if err != nil {
		// Row state unknown — do NOT touch the shared path: if the row is in
		// fact gone, these may be someone else's files.
		log.Printf("failed to delete image(%s) during compensation, leaving project files untouched: %v", image.Tag(), err)
		return
	}
	if !owned {
		log.Printf("image(%s) row already deleted by someone else during compensation; the project path may hold a newer create's files, leaving them", image.Tag())
		return
	}
	if err := u.imageService.CleanupProject(image); err != nil {
		log.Printf("failed to clean up project for image(%s) during compensation: %v", image.Tag(), err)
	}
}

func (u *usecase) deleteOwnImage(image *imageentity.Image, ownershipToken int) {
	// The reserve-failure path never wrote files, so the verdict is unused.
	if _, err := u.imageRepository.DeleteIfOwned(ownershipToken); err != nil {
		log.Printf("failed to delete image(%s) during compensation: %v", image.Tag(), err)
	}
}
