package exportproject_test

import (
	"bytes"
	"errors"
	"testing"

	"judge-backend/internal/app/project_domain/project/application/bundle"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	exportproject "judge-backend/internal/app/project_domain/project/application/usecase/project/export_project"
	"judge-backend/internal/app/project_domain/project/mocks"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/app/shared/types/zipfile"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type exportMocks struct {
	query   *mocks.ProjectExportQuery
	assets  *mocks.AssetQuery
	grading *mocks.GradingProjectService
	zip     *sharedmocks.ZipService
	clock   *mocks.Clock
	// logger is strict: no expectation is registered by default, so ANY log
	// call fails the test. The "no grading files" case relies on this — a
	// legitimate absence must leave nothing in the log.
	logger *sharedmocks.Logger
}

func newExportUsecase(t *testing.T) (exportproject.Usecase, *exportMocks) {
	m := &exportMocks{
		query:   mocks.NewProjectExportQuery(t),
		assets:  mocks.NewAssetQuery(t),
		grading: mocks.NewGradingProjectService(t),
		zip:     sharedmocks.NewZipService(t),
		clock:   mocks.NewClock(t),
		logger:  sharedmocks.NewLogger(t),
	}
	return exportproject.NewUsecase(m.query, m.assets, m.grading, m.zip, m.clock, m.logger), m
}

func minimalProject() *dto.ProjectExport {
	return &dto.ProjectExport{ID: 42, Name: "Minimal", ImageTag: "img:1", TestImageTag: "test:1"}
}

// captureZip records the file list handed to the zip service so the test can
// inspect what the bundle would contain.
func captureZip(m *exportMocks, into *[]zipfile.File) {
	m.zip.On("ZipFromFiles", mock.MatchedBy(func(files *[]zipfile.File) bool {
		*into = append([]zipfile.File(nil), (*files)...)
		return true
	})).Return(bytes.NewBufferString("zip"), nil).Once()
}

func definitionIn(t *testing.T, files []zipfile.File) *bundle.Project {
	t.Helper()
	for _, f := range files {
		if f.Path == bundle.ProjectFile {
			p, err := bundle.Unmarshal([]byte(f.Body))
			require.NoError(t, err)
			return p
		}
	}
	t.Fatalf("bundle has no %s", bundle.ProjectFile)
	return nil
}

func hasPath(files []zipfile.File, path string) bool {
	for _, f := range files {
		if f.Path == path {
			return true
		}
	}
	return false
}

// No grading files uploaded is a legitimate state: the export succeeds, the
// definition says hasGradingFiles=false, the zip carries no grading-files
// entry — and nothing is logged (the strict logger mock has no expectation).
func TestExport_NoGradingFiles_ExportsQuietly(t *testing.T) {
	uc, m := newExportUsecase(t)
	m.query.On("LoadProjectForExport", int32(42)).Return(minimalProject(), nil).Once()
	m.grading.On("DownloadGradingProject", int32(42)).Return(nil, gradingproject.ErrorGradingProjectNotFound).Once()
	m.clock.On("NowRFC3339").Return("2026-09-09T00:00:00Z").Once()
	var files []zipfile.File
	captureZip(m, &files)

	out, err := uc.Execute(&exportproject.Input{ProjectID: 42})

	require.NoError(t, err)
	assert.Equal(t, "minimal-42.zip", out.Filename)
	assert.False(t, definitionIn(t, files).HasGradingFiles)
	assert.False(t, hasPath(files, bundle.GradingFilesFile))
}

func TestExport_WithGradingFiles_IncludesThem(t *testing.T) {
	uc, m := newExportUsecase(t)
	m.query.On("LoadProjectForExport", int32(42)).Return(minimalProject(), nil).Once()
	m.grading.On("DownloadGradingProject", int32(42)).
		Return(&gradingproject.File{Name: "project-42.zip", Content: []byte("grading-zip-bytes"), Size: 17}, nil).Once()
	m.clock.On("NowRFC3339").Return("2026-09-09T00:00:00Z").Once()
	var files []zipfile.File
	captureZip(m, &files)

	_, err := uc.Execute(&exportproject.Input{ProjectID: 42})

	require.NoError(t, err)
	assert.True(t, definitionIn(t, files).HasGradingFiles)
	require.True(t, hasPath(files, bundle.GradingFilesFile))
	for _, f := range files {
		if f.Path == bundle.GradingFilesFile {
			assert.Equal(t, "grading-zip-bytes", f.Body)
		}
	}
}

// Grading files that exist but cannot be read must fail the export: a bundle
// silently missing them would import into a problem that grades differently.
// The cause is logged (the usecase's fail convention) and the caller gets the
// opaque internal error, never a zip.
func TestExport_GradingFilesUnreadable_FailsLoudly(t *testing.T) {
	uc, m := newExportUsecase(t)
	m.query.On("LoadProjectForExport", int32(42)).Return(minimalProject(), nil).Once()
	m.grading.On("DownloadGradingProject", int32(42)).Return(nil, errors.New("permission denied")).Once()
	m.logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Once()
	// No ZipFromFiles / NowRFC3339 expectations: reaching them would fail.

	out, err := uc.Execute(&exportproject.Input{ProjectID: 42})

	assert.Nil(t, out)
	assert.ErrorIs(t, err, exportproject.ErrorInternalServer)
}
