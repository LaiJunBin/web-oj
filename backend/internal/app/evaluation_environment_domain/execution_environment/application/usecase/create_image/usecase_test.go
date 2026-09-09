package createimage_test

import (
	"archive/zip"
	"bytes"
	"errors"
	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"sync"
	"testing"

	testifymock "github.com/stretchr/testify/mock"
)

type File struct {
	Name, Body string
}

func CreateZipFile(files ...*File) *bytes.Buffer {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}

	err := w.Close()
	if err != nil {
		panic(err)
	}

	return buf
}

type CreateImageUsecaseMocks struct {
	imageRepository *mocks.ImageRepository
	imageService    *mocks.ImageService
	buildService    *mocks.BuildService
}

func NewCreateImageUsecaseMocks(t *testing.T) *CreateImageUsecaseMocks {
	return &CreateImageUsecaseMocks{
		imageRepository: mocks.NewImageRepository(t),
		imageService:    mocks.NewImageService(t),
		buildService:    mocks.NewBuildService(t),
	}
}

func NewImageFactory() *factory.ImageFactory {
	storagePathProvider := &mocks.StoragePathProvider{}
	storagePathProvider.On("GetStoragePath").Return("storage")
	return factory.NewImageFactory(storagePathProvider)
}

func NewCreateImageUsecase(mocks *CreateImageUsecaseMocks) createimage.Usecase {
	factory := NewImageFactory()
	return createimage.NewUsecase(factory, mocks.imageRepository, mocks.imageService, mocks.buildService)
}

func NewPendingBuild(tag string) *build.Build {
	return build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew(tag),
		Status: buildstatus.PENDING,
		Batch:  batch.MustNew(1),
	})
}

func NewImage(config *image.Config) *image.Image {
	factory := NewImageFactory()
	image, _ := factory.CreateImage(config.Tag, config.MountPath, config.Description)
	return image
}

func TestCreateImageUsecase(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		t.Run("empty tag", func(t *testing.T) {
			mocks := NewCreateImageUsecaseMocks(t)
			mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
			uc := NewCreateImageUsecase(mocks)
			input := &createimage.Input{
				Tag:            "",
				MountPath:      "/var/www/html",
				Description:    "description",
				ProjectZipFile: mockProject,
			}

			output, err := uc.Execute(input)

			if !errors.Is(err, createimage.ErrorTagIsRequired) {
				t.Fatalf("Expected error to be ErrInvalidInput, got %v", err)
			}

			if output != nil {
				t.Fatalf("Expected output to be nil, got %v", output)
			}
		})

		t.Run("invalid tag", func(t *testing.T) {
			mocks := NewCreateImageUsecaseMocks(t)
			mocks.imageRepository.On("FindByTag", "invalid_tag").Return(nil, nil)
			mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
			uc := NewCreateImageUsecase(mocks)
			input := &createimage.Input{
				Tag:            "invalid_tag",
				MountPath:      "/var/www/html",
				Description:    "description",
				ProjectZipFile: mockProject,
			}

			output, err := uc.Execute(input)

			if !errors.Is(err, createimage.ErrorInvalidTag) {
				t.Fatalf("Expected error to be ErrorInvalidTag, got %v", err)
			}

			if output != nil {
				t.Fatalf("Expected output to be nil, got %v", output)
			}
		})

		t.Run("empty zip project", func(t *testing.T) {
			mocks := NewCreateImageUsecaseMocks(t)
			uc := NewCreateImageUsecase(mocks)
			input := &createimage.Input{
				Tag:            "nginx:latest",
				MountPath:      "/var/www/html",
				Description:    "description",
				ProjectZipFile: nil,
			}

			output, err := uc.Execute(input)

			if !errors.Is(err, createimage.ErrorProjectZipIsRequired) {
				t.Fatalf("Expected error to be ErrInvalidInput, got %v", err)
			}

			if output != nil {
				t.Fatalf("Expected output to be nil, got %v", output)
			}
		})

		t.Run("duplicate tag", func(t *testing.T) {
			mockImage := NewImage(&image.Config{
				Tag:         "nginx:latest",
				Description: "description",
				MountPath:   "/var/www/html",
			})
			mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
			mocks := NewCreateImageUsecaseMocks(t)
			mocks.imageRepository.On("FindByTag", mockImage.Tag().String()).Return(mockImage, nil)
			uc := NewCreateImageUsecase(mocks)
			input := &createimage.Input{
				Tag:            mockImage.Tag().String(),
				MountPath:      "/var/www/html",
				Description:    mockImage.Description(),
				ProjectZipFile: mockProject,
			}

			output, err := uc.Execute(input)

			if !errors.Is(err, createimage.ErrorImageAlreadyExists) {
				t.Fatalf("Expected error to be ErrImageAlreadyExists, got %v", err)
			}

			if output != nil {
				t.Fatalf("Expected output to be nil, got %v", output)
			}
		})

		t.Run("invalid mount path", func(t *testing.T) {
			mockTag := "nginx:latest"
			mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
			mocks := NewCreateImageUsecaseMocks(t)
			mocks.imageRepository.On("FindByTag", mockTag).Return(nil, nil)
			uc := NewCreateImageUsecase(mocks)
			input := &createimage.Input{
				Tag:            mockTag,
				MountPath:      "",
				ProjectZipFile: mockProject,
			}

			output, err := uc.Execute(input)

			if !errors.Is(err, createimage.ErrorInvalidMountPath) {
				t.Fatalf("Expected error to be ErrInvalidInput, got %v", err)
			}

			if output != nil {
				t.Fatalf("Expected output to be nil, got %v", output)
			}
		})
	})
	// The two-phase create contract: CreateExclusive (DB tag ownership) →
	// ReserveBuild → SetupProject → EnqueueReservedBuild, each failure
	// compensating exactly what this request created.

	// Only the DB winner may touch the shared project path: with strict
	// mocks, SetupProject is allowed exactly ONCE — a second call (the old
	// setup-before-ownership order let every racer write the same
	// directory) panics the mock. Exactly one create succeeds; the loser
	// gets ErrorImageAlreadyExists and no compensation runs for it.
	t.Run("concurrent same-tag creates: only the DB winner sets up the project", func(t *testing.T) {
		mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
		mocks := NewCreateImageUsecaseMocks(t)
		reserved := NewPendingBuild("nginx:latest")

		mocks.imageRepository.On("FindByTag", "nginx:latest").Return(nil, nil).Twice()
		// The unique index arbitrates: whichever goroutine's insert lands
		// first wins, the other gets the duplicate verdict.
		mocks.imageRepository.On("CreateExclusive", testifymock.Anything).Return(7, nil).Once()
		mocks.imageRepository.On("CreateExclusive", testifymock.Anything).Return(0, repository.ErrImageTagAlreadyExists).Once()
		mocks.buildService.On("ReserveBuild", testifymock.Anything).Return(reserved, nil).Once()
		mocks.imageService.On("SetupProject", testifymock.Anything, mockProject).Return(nil).Once()
		mocks.buildService.On("EnqueueReservedBuild", reserved, testifymock.Anything, &buildservice.BuildOptions{}).Return(nil).Once()

		uc := NewCreateImageUsecase(mocks)
		input := func() *createimage.Input {
			return &createimage.Input{
				Tag:            "nginx:latest",
				MountPath:      "/var/www/html",
				Description:    "description",
				ProjectZipFile: mockProject,
			}
		}

		var wg sync.WaitGroup
		start := make(chan struct{})
		errs := make([]error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-start
				_, errs[idx] = uc.Execute(input())
			}(i)
		}
		close(start)
		wg.Wait()

		winners, losers := 0, 0
		for _, err := range errs {
			switch err {
			case nil:
				winners++
			case createimage.ErrorImageAlreadyExists:
				losers++
			default:
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if winners != 1 || losers != 1 {
			t.Fatalf("expected exactly one winner and one ErrorImageAlreadyExists loser, got %d winners / %d losers", winners, losers)
		}
	})

	// Reserve failure: setup must never run (no expectation registered) and
	// the image row this request created must be deleted — otherwise a
	// retry forever gets ErrorImageAlreadyExists for an image with no build.
	t.Run("reserve failure deletes the created image and never sets up", func(t *testing.T) {
		mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
		mocks := NewCreateImageUsecaseMocks(t)

		mocks.imageRepository.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		mocks.imageRepository.On("CreateExclusive", testifymock.Anything).Return(7, nil).Once()
		mocks.buildService.On("ReserveBuild", testifymock.Anything).Return(nil, repository.ErrorBuildAlreadyInProgress).Once()
		mocks.imageRepository.On("DeleteIfOwned", 7).Return(true, nil).Once()
		// No SetupProject / CleanupProject / ReleaseReservedBuild
		// expectations: nothing was set up and no reservation was acquired.

		uc := NewCreateImageUsecase(mocks)
		_, err := uc.Execute(&createimage.Input{
			Tag: "nginx:latest", MountPath: "/var/www/html", Description: "d", ProjectZipFile: mockProject,
		})
		if err != createimage.ErrorInternalServer {
			t.Fatalf("expected ErrorInternalServer, got %v", err)
		}
	})

	// Setup failure: release the reservation, clean the partially-written
	// project, delete the image row.
	t.Run("setup failure releases reservation, cleans project and deletes image", func(t *testing.T) {
		mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
		mocks := NewCreateImageUsecaseMocks(t)
		reserved := NewPendingBuild("nginx:latest")

		mocks.imageRepository.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		mocks.imageRepository.On("CreateExclusive", testifymock.Anything).Return(7, nil).Once()
		mocks.buildService.On("ReserveBuild", testifymock.Anything).Return(reserved, nil).Once()
		mocks.imageService.On("SetupProject", testifymock.Anything, mockProject).Return(errors.New("unzip exploded")).Once()
		mocks.buildService.On("ReleaseReservedBuild", reserved).Return().Once()
		mocks.imageService.On("CleanupProject", testifymock.Anything).Return(nil).Once()
		mocks.imageRepository.On("DeleteIfOwned", 7).Return(true, nil).Once()

		uc := NewCreateImageUsecase(mocks)
		_, err := uc.Execute(&createimage.Input{
			Tag: "nginx:latest", MountPath: "/var/www/html", Description: "d", ProjectZipFile: mockProject,
		})
		if err != createimage.ErrorFailedToSetup {
			t.Fatalf("expected ErrorFailedToSetup, got %v", err)
		}
	})

	// Enqueue failure: same full compensation — and the error surfaces
	// instead of being logged away.
	t.Run("enqueue failure releases reservation, cleans project and deletes image", func(t *testing.T) {
		mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
		mocks := NewCreateImageUsecaseMocks(t)
		reserved := NewPendingBuild("nginx:latest")

		mocks.imageRepository.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		mocks.imageRepository.On("CreateExclusive", testifymock.Anything).Return(7, nil).Once()
		mocks.buildService.On("ReserveBuild", testifymock.Anything).Return(reserved, nil).Once()
		mocks.imageService.On("SetupProject", testifymock.Anything, mockProject).Return(nil).Once()
		mocks.buildService.On("EnqueueReservedBuild", reserved, testifymock.Anything, &buildservice.BuildOptions{}).Return(errors.New("queue exploded")).Once()
		mocks.buildService.On("ReleaseReservedBuild", reserved).Return().Once()
		mocks.imageService.On("CleanupProject", testifymock.Anything).Return(nil).Once()
		mocks.imageRepository.On("DeleteIfOwned", 7).Return(true, nil).Once()

		uc := NewCreateImageUsecase(mocks)
		_, err := uc.Execute(&createimage.Input{
			Tag: "nginx:latest", MountPath: "/var/www/html", Description: "d", ProjectZipFile: mockProject,
		})
		if err != createimage.ErrorInternalServer {
			t.Fatalf("expected ErrorInternalServer, got %v", err)
		}
	})
	t.Run("success", func(t *testing.T) {
		mockImage := NewImage(&image.Config{
			Tag:         "nginx:latest",
			Description: "description",
			MountPath:   "/var/www/html",
		})
		mockProject := CreateZipFile(&File{Name: "Dockerfile", Body: "FROM nginx:latest\n"})
		mocks := NewCreateImageUsecaseMocks(t)
		reserved := NewPendingBuild("nginx:latest")

		var order []string
		mocks.imageRepository.On("FindByTag", mockImage.Tag().String()).Return(nil, nil)
		mocks.imageRepository.On("CreateExclusive", mockImage).Return(7, nil).Run(func(testifymock.Arguments) {
			order = append(order, "save")
		}).Once()
		mocks.buildService.On("ReserveBuild", mockImage).Return(reserved, nil).Run(func(testifymock.Arguments) {
			order = append(order, "reserve")
		}).Once()
		mocks.imageService.On("SetupProject", mockImage, mockProject).Return(nil).Run(func(testifymock.Arguments) {
			order = append(order, "setup")
		}).Once()
		mocks.buildService.On("EnqueueReservedBuild", reserved, mockImage, &buildservice.BuildOptions{}).Return(nil).Run(func(testifymock.Arguments) {
			order = append(order, "enqueue")
		}).Once()
		defer func() {
			want := []string{"save", "reserve", "setup", "enqueue"}
			if len(order) != len(want) {
				t.Fatalf("expected call order %v, got %v", want, order)
			}
			for i := range want {
				if order[i] != want[i] {
					t.Fatalf("expected call order %v, got %v", want, order)
				}
			}
		}()
		uc := NewCreateImageUsecase(mocks)
		input := &createimage.Input{
			Tag:            mockImage.Tag().String(),
			MountPath:      "/var/www/html",
			Description:    mockImage.Description(),
			ProjectZipFile: mockProject,
		}

		output, err := uc.Execute(input)

		if err != nil {
			t.Fatalf("Execute should not return an error, but got %v", err)
		}

		if output == nil {
			t.Fatalf("Expected output to be not nil, got nil")
		}
	})
}
