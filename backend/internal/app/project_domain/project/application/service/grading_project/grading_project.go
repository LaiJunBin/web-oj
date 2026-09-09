package gradingproject

import (
	"bytes"
	"errors"
)

type File struct {
	Name    string
	Content []byte
	Size    int64
}

var (
	ErrorInvalidZipProject  = errors.New("invalid zip project")
	ErrorPathOutsideProject = errors.New("path escapes project files directory")
	// ErrorGradingProjectNotFound: no grading project has been uploaded for
	// this project. A legitimate state, not a failure — callers decide
	// whether it is a 404 or simply "nothing to include".
	ErrorGradingProjectNotFound = errors.New("grading project not found")
)

type GradingProjectService interface {
	DownloadGradingProject(projectID int32) (*File, error)
	UploadGradingProject(projectID int32, zipProject *bytes.Buffer) error
	ReadFile(projectID int32, path string) (*bytes.Buffer, error)
	ListFile(projectID int32) ([]string, error)
}
