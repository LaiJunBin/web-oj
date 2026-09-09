package service

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	logger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/app/shared/types/submissiondynamicfield"
)

const (
	testStoragePath = "./storage"
	testDirPath     = "./storage/submissions/sub-1"
	testProjectPath = "./storage/submissions/sub-1/project"
	testArchivePath = "./storage/submissions/sub-1/project.zip"
)

func newSubmissionService(t *testing.T) (*sharedmocks.FileSystemService, *sharedmocks.ZipService, *SubmissionService) {
	t.Helper()

	storagePathProvider := judgemocks.NewStoragePathProvider(t)
	storagePathProvider.On("GetStoragePath").Return(testStoragePath).Maybe()

	fss := sharedmocks.NewFileSystemService(t)
	zip := sharedmocks.NewZipService(t)

	svc := NewSubmissionService(storagePathProvider, fss, zip, logger.NewMockLogger())

	return fss, zip, svc.(*SubmissionService)
}

func TestSubmissionService_SaveProject(t *testing.T) {
	project := bytes.NewBufferString("PK-pretend-this-is-a-zip")

	// Only the extracted tree was kept, so a submission that misbehaved in
	// production could not be reproduced afterwards — exactly the situation
	// the unwrap and dotfile bugs left us in. Keeping the archive the
	// contestant actually uploaded is what makes a post-mortem possible.
	t.Run("writes the uploaded archive alongside the extracted project", func(t *testing.T) {
		fss, zip, svc := newSubmissionService(t)

		fss.On("MkdirAll", testProjectPath).Return(nil).Once()
		fss.On("WriteFile", testArchivePath, project.Bytes()).Return(nil).Once()
		zip.On("UnzipToPath", project, testProjectPath, true).Return(nil).Once()

		assert.NoError(t, svc.SaveProject("sub-1", project))
	})

	// Order matters more than it looks: extraction is the step most likely
	// to fail, and a post-mortem needs the archive that caused it. Writing
	// the archive first means a failed extraction still leaves evidence.
	t.Run("writes the archive before extracting", func(t *testing.T) {
		fss, zip, svc := newSubmissionService(t)

		archiveWritten := false

		fss.On("MkdirAll", testProjectPath).Return(nil).Once()
		fss.On("WriteFile", testArchivePath, project.Bytes()).
			Run(func(mock.Arguments) { archiveWritten = true }).
			Return(nil).Once()
		zip.On("UnzipToPath", project, testProjectPath, true).
			Run(func(mock.Arguments) {
				assert.True(t, archiveWritten, "the archive must be on disk before extraction is attempted")
			}).
			Return(nil).Once()

		assert.NoError(t, svc.SaveProject("sub-1", project))
	})

	t.Run("keeps the archive when extraction fails", func(t *testing.T) {
		fss, zip, svc := newSubmissionService(t)

		unzipErr := errors.New("bad archive")

		fss.On("MkdirAll", testProjectPath).Return(nil).Once()
		fss.On("WriteFile", testArchivePath, project.Bytes()).Return(nil).Once()
		zip.On("UnzipToPath", project, testProjectPath, true).Return(unzipErr).Once()

		assert.ErrorIs(t, svc.SaveProject("sub-1", project), unzipErr)

		// The archive write is asserted by mockery on cleanup: it was
		// expected Once and must have happened even though the call failed.
	})

	t.Run("does not extract when the archive cannot be written", func(t *testing.T) {
		fss, zip, svc := newSubmissionService(t)

		writeErr := errors.New("disk full")

		fss.On("MkdirAll", testProjectPath).Return(nil).Once()
		fss.On("WriteFile", testArchivePath, project.Bytes()).Return(writeErr).Once()

		assert.ErrorIs(t, svc.SaveProject("sub-1", project), writeErr)

		zip.AssertNotCalled(t, "UnzipToPath", mock.Anything, mock.Anything, mock.Anything)
	})

	// The existing guard: an empty ID would target storage/submissions
	// itself. Adding a second write under that path makes it more important,
	// not less.
	t.Run("refuses an empty submission ID before touching the filesystem", func(t *testing.T) {
		fss, zip, svc := newSubmissionService(t)

		assert.Error(t, svc.SaveProject("", project))

		fss.AssertNotCalled(t, "MkdirAll", mock.Anything)
		fss.AssertNotCalled(t, "WriteFile", mock.Anything, mock.Anything)
		zip.AssertNotCalled(t, "UnzipToPath", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestSubmissionService_CleanupSubmissionFiles(t *testing.T) {
	// Cleanup runs on the failure paths of submission creation, which are
	// precisely the submissions worth reproducing. Removing the whole
	// directory would take the uploaded archive with it and leave nothing to
	// investigate, so the archive is deliberately spared: a directory
	// holding only project.zip means "this submission failed to land".
	t.Run("removes the extracted files but keeps the uploaded archive", func(t *testing.T) {
		fss, _, svc := newSubmissionService(t)

		fss.On("RemoveAll", testProjectPath).Return(nil).Once()
		fss.On("RemoveAll", testDirPath+"/files").Return(nil).Once()
		fss.On("RemoveAll", testDirPath+"/metadata.json").Return(nil).Once()

		svc.CleanupSubmissionFiles("sub-1")

		fss.AssertNotCalled(t, "RemoveAll", testDirPath)
		fss.AssertNotCalled(t, "RemoveAll", testArchivePath)
	})

	t.Run("refuses an empty submission ID", func(t *testing.T) {
		fss, _, svc := newSubmissionService(t)

		svc.CleanupSubmissionFiles("")

		fss.AssertNotCalled(t, "RemoveAll", mock.Anything)
	})
}

func TestSubmissionService_SaveDynamicFields(t *testing.T) {
	const filesDir = "./storage/submissions/sub-1/files"
	const metadataPath = "./storage/submissions/sub-1/metadata.json"

	// A malicious project bundle can name a file field "../../config.yaml"; the
	// contestant then supplies its content. The write must be contained to the
	// submission's files directory and never reach the escaping path.
	t.Run("rejects a file field key that escapes the files directory", func(t *testing.T) {
		fss, _, svc := newSubmissionService(t)

		fss.On("MkdirAll", filesDir).Return(nil).Once()

		field, err := submissiondynamicfield.NewFile("../../evil", "meta,QUJD")
		assert.NoError(t, err)

		err = svc.SaveDynamicFields("sub-1", []*submissiondynamicfield.SubmissionDynamicField{field})
		assert.Error(t, err)

		fss.AssertNotCalled(t, "WriteFile", mock.Anything, mock.Anything)
	})

	t.Run("writes a contained file field and its metadata", func(t *testing.T) {
		fss, _, svc := newSubmissionService(t)

		fss.On("MkdirAll", filesDir).Return(nil).Once()
		fss.On("WriteFile", "storage/submissions/sub-1/files/answer", []byte("ABC")).Return(nil).Once()
		fss.On("WriteFile", metadataPath, mock.Anything).Return(nil).Once()

		field, err := submissiondynamicfield.NewFile("answer", "meta,QUJD")
		assert.NoError(t, err)

		assert.NoError(t, svc.SaveDynamicFields("sub-1", []*submissiondynamicfield.SubmissionDynamicField{field}))
	})
}
