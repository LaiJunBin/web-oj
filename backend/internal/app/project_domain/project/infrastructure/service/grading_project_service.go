package service

import (
	"bytes"
	"fmt"
	sharedservice "judge-backend/internal/app/shared/service"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	"judge-backend/internal/app/project_domain/project/domain/provider"
	"path/filepath"
	"strings"
)

type GradingProjectService struct {
	storagePath string
	fileService sharedservice.FileSystemService
	zipService  sharedservice.ZipService
}

// NewGradingProjectService returns an error instead of panicking when the
// storage directory cannot be created — consistent with this service's own
// methods, which already return errors for the same MkdirAll failure.
func NewGradingProjectService(
	storagePathProvider provider.StoragePathProvider,
	fileService sharedservice.FileSystemService,
	zipService sharedservice.ZipService,
) (gradingproject.GradingProjectService, error) {
	storagePath := fmt.Sprintf("%s/projects", storagePathProvider.GetStoragePath())
	if err := fileService.MkdirAll(storagePath); err != nil {
		return nil, fmt.Errorf("create grading project storage dir %s: %w", storagePath, err)
	}

	return &GradingProjectService{
		storagePath: storagePath,
		fileService: fileService,
		zipService:  zipService,
	}, nil
}

func (s *GradingProjectService) getProjectDirPath(projectID int32) string {
	return fmt.Sprintf("%s/%d", s.storagePath, projectID)
}

func (s *GradingProjectService) getProjectPath(projectID int32) string {
	return fmt.Sprintf("%s/%s", s.getProjectDirPath(projectID), "project.zip")
}

func (s *GradingProjectService) getProjectFilesPath(projectID int32) string {
	return fmt.Sprintf("%s/%s", s.getProjectDirPath(projectID), "files")
}

func (s *GradingProjectService) DownloadGradingProject(projectID int32) (*gradingproject.File, error) {
	path := s.getProjectPath(projectID)
	// A project with no uploaded grading files is an ordinary state, so it
	// is probed first: ReadFile logs every failure at error level (rightly —
	// for its other callers a missing file IS an anomaly), and a legitimate
	// absence must not show up in the log as one. Real read failures
	// (permissions, I/O) still go through ReadFile and keep its log line.
	if !s.fileService.Exists(path) {
		return nil, gradingproject.ErrorGradingProjectNotFound
	}
	file, err := s.fileService.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &gradingproject.File{
		Name:    fmt.Sprintf("project-%d.zip", projectID),
		Content: file,
		Size:    int64(len(file)),
	}, nil
}

func (s *GradingProjectService) UploadGradingProject(projectID int32, zipProject *bytes.Buffer) error {
	if err := s.zipService.CheckZip(zipProject); err != nil {
		return gradingproject.ErrorInvalidZipProject
	}

	oldProjectPath := s.getProjectPath(projectID)
	tmpProjectPath := fmt.Sprintf("%s/%d.old", s.storagePath, projectID)
	projectDirPath := s.getProjectDirPath(projectID)
	projectFilesPath := s.getProjectFilesPath(projectID)

	rollback := func() {
		s.fileService.RemoveAll(projectDirPath)
		s.fileService.Move(tmpProjectPath, oldProjectPath)
	}

	if s.fileService.Exists(oldProjectPath) {
		if err := s.fileService.Move(oldProjectPath, tmpProjectPath); err != nil {
			return err
		}
	}

	if err := s.fileService.RemoveAll(projectDirPath); err != nil {
		s.fileService.Move(tmpProjectPath, oldProjectPath)
		return err
	}

	if err := s.fileService.MkdirAll(projectDirPath); err != nil {
		rollback()
		return err
	}

	if err := s.fileService.WriteFile(s.getProjectPath(projectID), zipProject.Bytes()); err != nil {
		rollback()
		return err
	}

	if err := s.fileService.MkdirAll(projectFilesPath); err != nil {
		rollback()
		return err
	}

	if err := s.zipService.UnzipToPath(zipProject, projectFilesPath, true); err != nil {
		rollback()
		return err
	}

	s.fileService.RemoveAll(tmpProjectPath)
	return nil
}

func (s *GradingProjectService) ReadFile(projectID int32, path string) (*bytes.Buffer, error) {
	filePath, err := containedPath(s.getProjectFilesPath(projectID), path)
	if err != nil {
		return nil, err
	}

	file, err := s.fileService.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(file), nil
}

// containedPath joins base and rel, then rejects the result unless it stays
// inside base. rel is caller-supplied (a base64-decoded path from the API),
// so without this check a "../" segment could escape the project's files
// directory and read arbitrary files reachable by the server process.
func containedPath(base string, rel string) (string, error) {
	cleanBase := filepath.Clean(base)
	joined := filepath.Join(cleanBase, rel)
	if joined != cleanBase && !strings.HasPrefix(joined, cleanBase+string(filepath.Separator)) {
		return "", gradingproject.ErrorPathOutsideProject
	}
	return joined, nil
}

func (s *GradingProjectService) ListFile(projectID int32) ([]string, error) {
	if !s.fileService.Exists(s.getProjectFilesPath(projectID)) {
		return []string{}, nil
	}

	files, err := s.fileService.ListFileRecursive(s.getProjectFilesPath(projectID))
	if err != nil {
		return nil, err
	}

	prefix := s.getProjectFilesPath(projectID)
	for i, file := range files {
		files[i] = file[len(prefix)-1:]
	}

	return files, nil
}
