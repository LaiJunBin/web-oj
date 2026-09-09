package createimage_test

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"

	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	infrarepo "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/repository"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/image_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	filesystem "judge-backend/internal/app/shared/infrastructure/services/file_system"
	zipservice "judge-backend/internal/app/shared/infrastructure/services/zip_service"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/projectpath"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// A LATE create-compensation must not remove a re-created image's project
// files. (Since rename-on-delete freed
// soft-deleted tags on MySQL's non-partial unique index, this delete→recreate
// interleaving is reachable on MySQL too, not only under sqlite's partial
// index; the fence proven here matters on every backend.) Real filesystem, real sqlite rows and a
// real image service on purpose — the defect is which files a tag-derived
// path points at by the time the compensation runs, which mocks cannot show.
//
// Interleaving (driven through a gate inside the mocked build service):
//
//	A: create → row + files written → enqueue blocks on the gate
//	    user deletes A's image (row gone, tag free)
//	    B: creates the same tag → new row + writes ITS files to the SAME path
//	A: enqueue fails → compensate → DeleteIfOwned(tokenA) says NOT owned
//	    → A must leave the path alone: B's row AND B's files survive.
func TestCreateImage_LateCompensationLeavesRecreatedImageFiles(t *testing.T) {
	const tag = "fencing-late-compensation:latest"
	storageRel := "tmp/fencing_late_compensation_test"
	storageAbs := path.Join(projectpath.Root, storageRel)
	require.NoError(t, os.MkdirAll(storageAbs, 0755))
	defer os.RemoveAll(storageAbs)

	// Real sqlite-backed repository (same construction as the repository
	// layer's ownership tests).
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Image{}))
	repoPathProvider := &mocks.StoragePathProvider{}
	repoPathProvider.On("GetStoragePath").Return(storageRel)
	realRepo := infrarepo.NewImageRepository(db, repoPathProvider)

	// Real image service over the real filesystem.
	log := logger.NewMockLogger()
	fss := filesystem.NewService(log)
	zips := zipservice.NewService(log, zipservice.Limits{})
	buildService := mocks.NewBuildService(t)
	realImageService := imageservice.NewService(realRepo, buildService, fss, zips)

	// The factory joins projectpath.Root itself, so it gets the RELATIVE path.
	factoryPathProvider := &mocks.StoragePathProvider{}
	factoryPathProvider.On("GetStoragePath").Return(storageRel)
	imageFactory := factory.NewImageFactory(factoryPathProvider)

	uc := createimage.NewUsecase(imageFactory, realRepo, realImageService, buildService)

	// Gate inside A's enqueue: A pauses with its row and files written.
	reserved := NewPendingBuild(tag)
	aEntered := make(chan struct{})
	aRelease := make(chan struct{})
	buildService.On("ReserveBuild", testifymock.Anything).Return(reserved, nil).Once()
	buildService.On("EnqueueReservedBuild", reserved, testifymock.Anything, testifymock.Anything).
		Run(func(testifymock.Arguments) {
			close(aEntered)
			<-aRelease
		}).Return(errors.New("enqueue exploded")).Once()
	buildService.On("ReleaseReservedBuild", reserved).Return().Once()

	aDone := make(chan error, 1)
	go func() {
		_, execErr := uc.Execute(&createimage.Input{
			Tag:            tag,
			MountPath:      "/var/www/html",
			Description:    "A's image",
			ProjectZipFile: CreateZipFile(&File{Name: "Dockerfile", Body: "FROM a\n"}),
		})
		aDone <- execErr
	}()

	select {
	case <-aEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("A never reached enqueue")
	}

	// The user deletes A's image while A is still in flight. (Delete resolves
	// the row by tag, which at this moment is still A's row.)
	imageAEquivalent, err := imageFactory.CreateImage(tag, "/var/www/html", "A's image")
	require.NoError(t, err)
	require.NoError(t, realRepo.Delete(imageAEquivalent))

	// B re-creates the same tag: new row, and ITS files at the same
	// tag-derived path (SetupProject wipes whatever was there first).
	imageB, err := imageFactory.CreateImage(tag, "/var/www/html", "B's image")
	require.NoError(t, err)
	_, err = realRepo.CreateExclusive(imageB)
	require.NoError(t, err)
	require.NoError(t, realImageService.SetupProject(imageB, CreateZipFile(&File{Name: "Dockerfile", Body: "FROM b\n"})))

	bDockerfile := path.Join(imageB.Project().StoragePath(), "Dockerfile")
	content, err := os.ReadFile(bDockerfile)
	require.NoError(t, err)
	require.Equal(t, "FROM b\n", string(content), "precondition: B's files are in place")

	// Release A: its enqueue fails, compensation runs against a tag that now
	// belongs to B.
	close(aRelease)
	select {
	case execErr := <-aDone:
		require.ErrorIs(t, execErr, createimage.ErrorInternalServer)
	case <-time.After(5 * time.Second):
		t.Fatal("A never finished")
	}

	// The verdict-fenced compensation must have left B alone.
	content, err = os.ReadFile(bDockerfile)
	require.NoError(t, err, "A's late compensation removed B's project files")
	require.Equal(t, "FROM b\n", string(content))

	current, err := realRepo.FindByTag(tag)
	require.NoError(t, err)
	require.NotNil(t, current, "B's row must survive A's compensation")
	require.Equal(t, "B's image", current.Description())
}
