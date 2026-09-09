package service

import (
	"errors"
	"testing"

	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
)

// TestContainedPath is a regression test for A2: GradingProjectService.ReadFile
// used to build the filesystem path by string-concatenating the caller-supplied
// (base64-decoded) path with no containment check, letting "../" segments
// escape the project's files directory and read arbitrary files reachable by
// the server process.
func TestContainedPath(t *testing.T) {
	base := "/data/projects/1/files"

	t.Run("plain relative path stays inside base", func(t *testing.T) {
		got, err := containedPath(base, "sub/dir/file.txt")
		assert.NoError(t, err)
		assert.Equal(t, "/data/projects/1/files/sub/dir/file.txt", got)
	})

	t.Run("parent traversal is rejected", func(t *testing.T) {
		_, err := containedPath(base, "../../../etc/passwd")
		assert.ErrorIs(t, err, gradingproject.ErrorPathOutsideProject)
	})

	t.Run("a leading slash in rel does not escape base", func(t *testing.T) {
		// Unlike e.g. Python's os.path.join, Go's filepath.Join does not let
		// a leading "/" in a later argument replace the base — it always
		// joins-and-cleans, so this nests under base rather than escaping.
		got, err := containedPath(base, "/etc/passwd")
		assert.NoError(t, err)
		assert.Equal(t, "/data/projects/1/files/etc/passwd", got)
	})

	t.Run("traversal that returns exactly to base is rejected", func(t *testing.T) {
		// filepath.Join(base, "..") cleans to the parent of base, not base
		// itself, so this must be rejected too.
		_, err := containedPath(base, "..")
		assert.ErrorIs(t, err, gradingproject.ErrorPathOutsideProject)
	})

	t.Run("a sneaky prefix-sibling directory is rejected", func(t *testing.T) {
		// "/data/projects/1/files-evil/secret" has the string "files" as a
		// prefix but is a SIBLING directory, not a descendant — a naive
		// strings.HasPrefix(joined, base) check (without the separator)
		// would wrongly accept this.
		_, err := containedPath(base, "../files-evil/secret")
		assert.ErrorIs(t, err, gradingproject.ErrorPathOutsideProject)
	})

	t.Run("empty relative path resolves to base itself", func(t *testing.T) {
		got, err := containedPath(base, "")
		assert.NoError(t, err)
		assert.Equal(t, base, got)
	})
}

func TestGradingProjectService_ReadFile_RejectsPathTraversal(t *testing.T) {
	fss := sharedmocks.NewFileSystemService(t)
	zip := sharedmocks.NewZipService(t)
	s := &GradingProjectService{
		storagePath: "/data/projects",
		fileService: fss,
		zipService:  zip,
	}

	// fss.ReadFile must NEVER be called for a traversal attempt — no
	// expectation is registered, so testify fails the test if it is.
	_, err := s.ReadFile(1, "../../../../etc/passwd")
	assert.ErrorIs(t, err, gradingproject.ErrorPathOutsideProject)
}

// A project with no uploaded grading files is an ordinary state. Download
// must report it with the sentinel and must not reach ReadFile, whose
// error-level log is right for its other callers but would brand this
// legitimate absence as a failure in the operator's log.
func TestGradingProjectService_Download_MissingIsNotFoundAndNeverReads(t *testing.T) {
	fss := sharedmocks.NewFileSystemService(t)
	s := &GradingProjectService{
		storagePath: "/data/projects",
		fileService: fss,
		zipService:  sharedmocks.NewZipService(t),
	}
	fss.On("Exists", "/data/projects/7/project.zip").Return(false).Once()
	// No ReadFile expectation: a call would fail the test.

	file, err := s.DownloadGradingProject(7)

	assert.Nil(t, file)
	assert.ErrorIs(t, err, gradingproject.ErrorGradingProjectNotFound)
}

func TestGradingProjectService_Download_ExistingIsRead(t *testing.T) {
	fss := sharedmocks.NewFileSystemService(t)
	s := &GradingProjectService{
		storagePath: "/data/projects",
		fileService: fss,
		zipService:  sharedmocks.NewZipService(t),
	}
	fss.On("Exists", "/data/projects/7/project.zip").Return(true).Once()
	fss.On("ReadFile", "/data/projects/7/project.zip").Return([]byte("zip-bytes"), nil).Once()

	file, err := s.DownloadGradingProject(7)

	assert.NoError(t, err)
	assert.Equal(t, "project-7.zip", file.Name)
	assert.Equal(t, []byte("zip-bytes"), file.Content)
	assert.EqualValues(t, 9, file.Size)
}

// A file that exists but cannot be read is a real failure: it goes through
// ReadFile (keeping its log line) and surfaces as-is, never as "not found".
func TestGradingProjectService_Download_ReadFailurePassesThrough(t *testing.T) {
	fss := sharedmocks.NewFileSystemService(t)
	s := &GradingProjectService{
		storagePath: "/data/projects",
		fileService: fss,
		zipService:  sharedmocks.NewZipService(t),
	}
	boom := errors.New("permission denied")
	fss.On("Exists", "/data/projects/7/project.zip").Return(true).Once()
	fss.On("ReadFile", "/data/projects/7/project.zip").Return(nil, boom).Once()

	_, err := s.DownloadGradingProject(7)

	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, gradingproject.ErrorGradingProjectNotFound)
}
