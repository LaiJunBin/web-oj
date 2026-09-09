package filesystem

import (
	"archive/tar"
	"bytes"
	"io/fs"
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"os"
	"path/filepath"
)

type Service struct {
	logger sharedtypes.Logger
}

func NewService(logger sharedtypes.Logger) sharedservice.FileSystemService {
	return &Service{
		logger: logger,
	}
}

func (f *Service) ListFileRecursive(path string) ([]string, error) {
	f.logger.Debugf("Listing files recursively in path: %s", path)
	files := []string{}
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		f.logger.Errorf("Failed to list files in path %s: %v", path, err)
		return nil, err
	}
	f.logger.Debugf("Successfully listed files in path: %s", path)
	return files, nil
}

func (f *Service) Exists(path string) bool {
	f.logger.Debugf("Checking existence of path: %s", path)
	_, err := os.Stat(path)
	exists := err == nil
	if exists {
		f.logger.Debugf("Path exists: %s", path)
	} else {
		f.logger.Debugf("Path does not exist: %s", path)
	}
	return exists
}

func (f *Service) MkdirAll(path string) error {
	f.logger.Debugf("Creating directories for path: %s", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		f.logger.Errorf("Failed to create directories for path %s: %v", path, err)
		return err
	}
	f.logger.Debugf("Successfully created directories for path: %s", path)
	return nil
}

func (f *Service) RemoveAll(path string) error {
	f.logger.Debugf("Removing all contents at path: %s", path)
	if err := os.RemoveAll(path); err != nil {
		f.logger.Errorf("Failed to remove contents at path %s: %v", path, err)
		return err
	}
	f.logger.Debugf("Successfully removed all contents at path: %s", path)
	return nil
}

func (f *Service) WriteFile(path string, data []byte) error {
	f.logger.Debugf("Writing file at path: %s", path)
	if err := os.WriteFile(path, data, 0644); err != nil {
		f.logger.Errorf("Failed to write file at path %s: %v", path, err)
		return err
	}
	f.logger.Debugf("Successfully wrote file at path: %s", path)
	return nil
}

func (f *Service) ReadFile(path string) ([]byte, error) {
	f.logger.Debugf("Reading file at path: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		f.logger.Errorf("Failed to read file at path %s: %v", path, err)
		return nil, err
	}
	f.logger.Debugf("Successfully read file at path: %s", path)
	return data, nil
}

func (f *Service) TarPath(path string) (*bytes.Buffer, error) {
	f.logger.Debugf("Creating tar archive for path: %s", path)
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	absPath, err := filepath.Abs(path)
	if err != nil {
		f.logger.Errorf("Failed to create tar archive for path %s: %v", path, err)
		return nil, err
	}

	err = filepath.Walk(absPath, func(file string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		absFile, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		header.Name = absFile[len(absPath):]
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := os.ReadFile(absFile)
			if err != nil {
				return err
			}

			if _, err := tw.Write(data); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		f.logger.Errorf("Failed to create tar archive for path %s: %v", path, err)
		return nil, err
	}

	if err := tw.Close(); err != nil {
		f.logger.Errorf("Failed to create tar archive for path %s: %v", path, err)
		return nil, err
	}

	f.logger.Debugf("Successfully created tar archive for path: %s", path)
	return buf, nil
}

func (f *Service) Move(src, dst string) error {
	f.logger.Debugf("Moving file from %s to %s", src, dst)
	if err := os.Rename(src, dst); err != nil {
		f.logger.Errorf("Failed to move file from %s to %s: %v", src, dst, err)
		return err
	}
	f.logger.Debugf("Successfully moved file from %s to %s", src, dst)
	return nil
}
