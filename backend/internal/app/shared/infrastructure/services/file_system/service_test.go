package filesystem_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"judge-backend/internal/app/shared/infrastructure/logger"
	filesystem "judge-backend/internal/app/shared/infrastructure/services/file_system"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/projectpath"
	"os"
	"path/filepath"
	"testing"
)

var (
	PATH_PREFIX = projectpath.Root + "/tmp/file_system_service_test"
)

func setup() {
	os.MkdirAll(PATH_PREFIX, 0755)
}

func teardown() {
	os.RemoveAll(PATH_PREFIX)
}

func NewFileSystemService(t *testing.T) sharedservice.FileSystemService {
	logger := logger.NewMockLogger()
	return filesystem.NewService(logger)
}

func createFile(path string) error {
	dirname := filepath.Dir(PATH_PREFIX + path)
	os.MkdirAll(dirname, 0755)
	f, err := os.Create(PATH_PREFIX + path)
	if err != nil {
		return err
	}
	return f.Close()
}

func assertFilePathIsAbsoulte(path string) {
	if !filepath.IsAbs(path) {
		panic(fmt.Sprintf("path %s is not absolute", path))
	}
}

func assertFileInList(file string, files []string) {
	rf := file[len(PATH_PREFIX):]
	for _, f := range files {
		if rf == f {
			return
		}
	}
	panic(fmt.Sprintf("file %s is not in list", file))
}

func TestFileSystemService_ListFileRecursive(t *testing.T) {
	setup()
	defer teardown()

	expectedFiles := []string{
		"/a/b/c/d.txt",
		"/a/b/c/e.txt",
		"/a/b/c/f.txt",
		"/a/b/c1.txt",
		"/a/b/c2.txt",
		"/a/b/c2/d.txt",
	}
	for _, file := range expectedFiles {
		createFile(file)
	}

	fss := NewFileSystemService(t)
	files, err := fss.ListFileRecursive(PATH_PREFIX)
	if err != nil {
		t.Error(err)
	}

	if len(files) != 6 {
		t.Errorf("expected 6 files, but got %d", len(files))
	}

	for _, file := range files {
		assertFilePathIsAbsoulte(file)
		assertFileInList(file, expectedFiles)
	}
}

func TestFileSystemService_MkdirAll(t *testing.T) {
	setup()
	defer teardown()

	fss := NewFileSystemService(t)
	err := fss.MkdirAll(PATH_PREFIX + "/a/b/c")
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(PATH_PREFIX + "/a/b/c")
	if err != nil {
		t.Error(err)
	}
}

func TestFileSystemService_RemoveAll(t *testing.T) {
	setup()
	defer teardown()

	os.MkdirAll(PATH_PREFIX+"/a/b/c/d", 0755)
	fss := NewFileSystemService(t)
	err := fss.RemoveAll(PATH_PREFIX + "/a/b")
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(PATH_PREFIX + "/a")
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(PATH_PREFIX + "/a/b")
	if err == nil {
		t.Error("directory /a/b should be removed")
	}
}

func TestFileSystemService_WriteFile(t *testing.T) {
	setup()
	defer teardown()

	fss := NewFileSystemService(t)
	err := fss.WriteFile(PATH_PREFIX+"/a.txt", []byte("hello"))
	if err != nil {
		t.Error(err)
	}

	data, err := os.ReadFile(PATH_PREFIX + "/a.txt")
	if err != nil {
		t.Error(err)
	}

	if string(data) != "hello" {
		t.Errorf("expected hello, but got %s", string(data))
	}
}

func TestFileSystemService_ReadFile(t *testing.T) {
	setup()
	defer teardown()

	os.WriteFile(PATH_PREFIX+"/a.txt", []byte("hello"), 0644)
	fss := NewFileSystemService(t)
	data, err := fss.ReadFile(PATH_PREFIX + "/a.txt")
	if err != nil {
		t.Error(err)
	}

	if string(data) != "hello" {
		t.Errorf("expected hello, but got %s", string(data))
	}
}

func TestFileSystemService_TarPath(t *testing.T) {
	setup()
	defer teardown()

	expectedFiles := []string{
		"/a/b/c/d.txt",
		"/a/b/c/e.txt",
		"/a/b/c/f.txt",
		"/a/b/c1.txt",
		"/a/b/c2.txt",
		"/a/b/c2/d.txt",
	}
	for _, file := range expectedFiles {
		createFile(file)
	}

	fss := NewFileSystemService(t)
	buf, err := fss.TarPath(PATH_PREFIX + "/a")
	if err != nil {
		t.Error(err)
	}

	if buf == nil {
		t.Error("expected buffer, but got nil")
	}

	r := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			t.Fatal(err)
		}

		if header.FileInfo().IsDir() {
			continue
		}

		if header.Name[:2] == "/a" {
			t.Fatalf("file %s should not have prefix /a", header.Name)
		}

		assertFileInList(PATH_PREFIX+"/a"+header.Name, expectedFiles)
	}
}

func TestFileSystemService_Move(t *testing.T) {
	t.Run("move file", func(t *testing.T) {
		setup()
		defer teardown()

		fss := NewFileSystemService(t)
		createFile("/a.txt")
		err := fss.Move(PATH_PREFIX+"/a.txt", PATH_PREFIX+"/b.txt")
		if err != nil {
			t.Error(err)
		}

		_, err = os.Stat(PATH_PREFIX + "/a.txt")
		if err == nil {
			t.Error("file /a.txt should be removed")
		}

		_, err = os.Stat(PATH_PREFIX + "/b.txt")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("move directory", func(t *testing.T) {
		setup()
		defer teardown()

		fss := NewFileSystemService(t)
		os.MkdirAll(PATH_PREFIX+"/a/b/c", 0755)
		err := fss.Move(PATH_PREFIX+"/a", PATH_PREFIX+"/b")
		if err != nil {
			t.Error(err)
		}

		_, err = os.Stat(PATH_PREFIX + "/a")
		if err == nil {
			t.Error("directory /a should be removed")
		}

		_, err = os.Stat(PATH_PREFIX + "/b")
		if err != nil {
			t.Error(err)
		}
	})
}
