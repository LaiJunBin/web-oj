package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"judge-backend/internal/app/project_domain/project/mocks"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A10 regression tests: a constructor that cannot create its storage
// directory must return an error (which fx reports as a clean startup
// failure naming the constructor) instead of panicking.

func TestNewAssetService_UncreatableStorageDirReturnsError(t *testing.T) {
	// A regular file where the parent directory should be makes MkdirAll fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("not a dir"), 0644))

	cfg := &config.Config{Server: &config.ServerConfig{StoragePath: blocker}}

	assert.NotPanics(t, func() {
		svc, err := NewAssetService(cfg)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}

func TestNewAssetService_CreatesDirAndSucceeds(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{Server: &config.ServerConfig{StoragePath: root}}

	svc, err := NewAssetService(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	info, statErr := os.Stat(filepath.Join(root, "assets"))
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}

func TestNewGradingProjectService_MkdirFailureReturnsError(t *testing.T) {
	pathProvider := mocks.NewStoragePathProvider(t)
	pathProvider.On("GetStoragePath").Return("storage")
	fss := sharedmocks.NewFileSystemService(t)
	fss.On("MkdirAll", "storage/projects").Return(errors.New("disk full"))

	assert.NotPanics(t, func() {
		svc, err := NewGradingProjectService(pathProvider, fss, sharedmocks.NewZipService(t))
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}

func TestNewGradingProjectService_Succeeds(t *testing.T) {
	pathProvider := mocks.NewStoragePathProvider(t)
	pathProvider.On("GetStoragePath").Return("storage")
	fss := sharedmocks.NewFileSystemService(t)
	fss.On("MkdirAll", "storage/projects").Return(nil)

	svc, err := NewGradingProjectService(pathProvider, fss, sharedmocks.NewZipService(t))
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}
