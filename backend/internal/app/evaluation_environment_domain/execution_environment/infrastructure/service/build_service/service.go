package buildservice

import (
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/events"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/queue"
)

type Service struct {
	eventBus     sharedtypes.EventBus
	queue        queue.BuildQueue
	fss          sharedservice.FileSystemService
	buildFactory factory.BuildFactory
	buildRepo    repository.BuildRepository
}

func NewService(
	eventBus sharedtypes.EventBus,
	queue queue.BuildQueue,
	fss sharedservice.FileSystemService,
	buildFactory factory.BuildFactory,
	buildRepo repository.BuildRepository,
) buildservice.BuildService {
	s := &Service{
		eventBus:     eventBus,
		queue:        queue,
		fss:          fss,
		buildFactory: buildFactory,
		buildRepo:    buildRepo,
	}

	return s
}

func (s *Service) EnqueueBuildJob(image *image.Image, options *buildservice.BuildOptions) error {
	if !s.fss.Exists(image.Project().StoragePath()) {
		return buildservice.ErrorImageProjectNotSet
	}

	build, err := s.ReserveBuild(image)
	if err != nil {
		return err
	}
	if err := s.EnqueueReservedBuild(build, image, options); err != nil {
		s.ReleaseReservedBuild(build)
		return err
	}
	return nil
}

// ReserveBuild claims the tag's exclusive pending slot; see the interface
// doc comment. SavePendingExclusive makes the existence check and the
// registration one atomic repository step, so two concurrent reservations
// for the same tag admit exactly one winner — the loser gets
// ErrorBuildAlreadyInProgress instead of silently overwriting the winner's
// pending build (which would also have handed both events the same batch
// number, since batch allocation reads only committed builds).
func (s *Service) ReserveBuild(image *image.Image) (*build.Build, error) {
	b, err := s.buildFactory.Create(image)
	if err != nil {
		return nil, err
	}
	if err := s.buildRepo.SavePendingExclusive(b); err != nil {
		return nil, err
	}
	return b, nil
}

// EnqueueReservedBuild publishes the build event for a reservation whose
// project preparation succeeded. The project-existence check runs HERE, at
// enqueue time, so a job can only enter the queue after setup completed.
func (s *Service) EnqueueReservedBuild(b *build.Build, image *image.Image, options *buildservice.BuildOptions) error {
	if !s.fss.Exists(image.Project().StoragePath()) {
		return buildservice.ErrorImageProjectNotSet
	}

	s.queue.Enqueue(event.NewBuildImageEvent(&event.BuildImageEventConfig{
		Tag:   b.Tag(),
		Batch: b.Batch().Value(),
		Path:  image.Project().StoragePath(),
		Options: &event.BuildImageOptions{
			NoCache: options.NoCache,
		},
	}))

	return nil
}

// ReleaseReservedBuild frees a failed reservation; identity-guarded by the
// repository, so releasing late can never evict a newer reservation.
func (s *Service) ReleaseReservedBuild(b *build.Build) {
	s.buildRepo.ReleasePending(b)
}

func (s *Service) Cancel(build *build.Build) {
	s.eventBus.Publish(event.NewCancelBuildEvent(build))
}
