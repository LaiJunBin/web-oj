package zipservice_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"judge-backend/internal/app/shared/infrastructure/logger"
	zipservice "judge-backend/internal/app/shared/infrastructure/services/zip_service"
	"judge-backend/internal/app/shared/types/zipfile"
	"judge-backend/internal/projectpath"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

var (
	PATH_PREFIX = projectpath.Root + "/tmp/zip_service_test"
)

func setup() {
	os.MkdirAll(PATH_PREFIX, 0755)
}

func teardown() {
	os.RemoveAll(PATH_PREFIX)
}

func TestZipService_CheckZip(t *testing.T) {
	logger := logger.NewMockLogger()
	t.Run("invalid zip", func(t *testing.T) {
		s := zipservice.NewService(logger, zipservice.Limits{})
		buf := new(bytes.Buffer)
		err := s.CheckZip(buf)
		if err == nil {
			t.Error("expected error, but got nil")
		}
	})

	t.Run("valid zip", func(t *testing.T) {
		s := zipservice.NewService(logger, zipservice.Limits{})
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		w.Close()
		err := s.CheckZip(buf)
		if err != nil {
			t.Error(err)
		}
	})
}

func buildZipEntries(t *testing.T, entries [][2]string) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for _, e := range entries {
		f, err := w.Create(e[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(e[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf
}

func TestZipService_ReadZipToMap(t *testing.T) {
	log := logger.NewMockLogger()

	t.Run("reads entries into a map", func(t *testing.T) {
		s := zipservice.NewService(log, zipservice.Limits{})
		buf := buildZipEntries(t, [][2]string{{"a.txt", "AAA"}, {"dir/b.txt", "BB"}})
		files, err := s.ReadZipToMap(buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(files["a.txt"]) != "AAA" || string(files["dir/b.txt"]) != "BB" {
			t.Fatalf("unexpected files: %v", files)
		}
	})

	t.Run("rejects too many entries", func(t *testing.T) {
		s := zipservice.NewService(log, zipservice.Limits{MaxEntries: 1})
		buf := buildZipEntries(t, [][2]string{{"a", "1"}, {"b", "2"}})
		if _, err := s.ReadZipToMap(buf); !errors.Is(err, zipservice.ErrorZipTooManyEntries) {
			t.Fatalf("expected too-many-entries, got %v", err)
		}
	})

	t.Run("rejects archives over the shared byte budget", func(t *testing.T) {
		s := zipservice.NewService(log, zipservice.Limits{MaxBytes: 4})
		// 2 + 3 = 5 bytes across two entries, over the 4-byte budget.
		buf := buildZipEntries(t, [][2]string{{"a", "12"}, {"b", "345"}})
		if _, err := s.ReadZipToMap(buf); !errors.Is(err, zipservice.ErrorZipTooLarge) {
			t.Fatalf("expected too-large, got %v", err)
		}
	})

	t.Run("rejects duplicate entry names", func(t *testing.T) {
		s := zipservice.NewService(log, zipservice.Limits{})
		buf := buildZipEntries(t, [][2]string{{"dup", "1"}, {"dup", "2"}})
		if _, err := s.ReadZipToMap(buf); !errors.Is(err, zipservice.ErrorZipDuplicateEntry) {
			t.Fatalf("expected duplicate error, got %v", err)
		}
	})
}

func testZipFromFiles(t *testing.T, fileMap map[string]string) {
	logger := logger.NewMockLogger()
	files := &[]zipfile.File{}
	for path, body := range fileMap {
		(*files) = append(*files, zipfile.File{Path: path, Body: body})
	}

	s := zipservice.NewService(logger, zipservice.Limits{})
	buf, err := s.ZipFromFiles(files)
	if err != nil {
		t.Error(err)
	}

	if buf == nil {
		t.Error("expected buffer, but got nil")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Error(err)
	}

	if len(zipReader.File) != len(fileMap) {
		t.Errorf("expected %d files, but got %d", len(fileMap), len(zipReader.File))
	}

	for _, f := range zipReader.File {
		if _, ok := fileMap[f.Name]; !ok {
			t.Errorf("file %s is not expected", f.Name)
		}

		rc, err := f.Open()
		if err != nil {
			t.Error(err)
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)
		if fileMap[f.Name] != buf.String() {
			t.Errorf("expected %s, but got %s", fileMap[f.Name], buf.String())
		}

		rc.Close()
	}
}

func TestZipService_ZipFromFiles(t *testing.T) {
	t.Run("empty files", func(t *testing.T) {
		testZipFromFiles(t, map[string]string{})
	})

	t.Run("single depth files", func(t *testing.T) {
		fileMap := map[string]string{
			"hello.txt": "hello",
			"world.txt": "world",
		}

		testZipFromFiles(t, fileMap)
	})

	t.Run("multi depth files", func(t *testing.T) {
		fileMap := map[string]string{
			"hi.txt":      "hi",
			"a/b/c/d.txt": "hello",
			"a/b/c/e.txt": "world",
		}

		testZipFromFiles(t, fileMap)
	})
}

func testZipFromPath(t *testing.T, fileMap map[string]string) {
	logger := logger.NewMockLogger()
	pathPrefix := PATH_PREFIX + "/zip_from_path"
	os.MkdirAll(pathPrefix, 0755)
	for path, body := range fileMap {
		dirname := filepath.Dir(pathPrefix + "/" + path)
		os.MkdirAll(dirname, 0755)
		err := os.WriteFile(pathPrefix+"/"+path, []byte(body), 0644)
		if err != nil {
			t.Error(err)
		}
	}

	s := zipservice.NewService(logger, zipservice.Limits{})
	buf, err := s.ZipFromPath(pathPrefix)
	if err != nil {
		t.Error(err)
	}

	if buf == nil {
		t.Fatalf("expected buffer, but got nil")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Error(err)
	}

	if len(zipReader.File) != len(fileMap) {
		t.Errorf("expected %d files, but got %d", len(fileMap), len(zipReader.File))
	}

	// Entry names are relative to the zipped root, so they can be compared
	// against fileMap directly. They used to be the absolute on-disk paths,
	// which this helper papered over by slicing the prefix back off.
	for _, f := range zipReader.File {
		file := f.Name
		if _, ok := fileMap[file]; !ok {
			t.Errorf("file %s is not expected", f.Name)
		}

		rc, err := f.Open()
		if err != nil {
			t.Error(err)
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)
		if fileMap[file] != buf.String() {
			t.Errorf("expected %s, but got %s", fileMap[file], buf.String())
		}

		rc.Close()
	}
}

func TestZipService_ZipFromPath(t *testing.T) {
	logger := logger.NewMockLogger()
	t.Run("path not exists", func(t *testing.T) {
		s := zipservice.NewService(logger, zipservice.Limits{})
		buf, err := s.ZipFromPath(PATH_PREFIX + "/not_exists")
		if buf != nil {
			t.Error("expected nil, but got buffer")
		}

		if err == nil {
			t.Error("expected error, but got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Run("single depth files", func(t *testing.T) {
			setup()
			defer teardown()

			fileMap := map[string]string{
				"hello.txt": "hello",
				"world.txt": "world",
			}

			testZipFromPath(t, fileMap)
		})

		t.Run("multi depth files", func(t *testing.T) {
			setup()
			defer teardown()

			fileMap := map[string]string{
				"hi.txt":      "hi",
				"a/b/c/d.txt": "hello",
				"a/b/c/e.txt": "world",
			}

			testZipFromPath(t, fileMap)
		})

		// Entry names must describe where a file sits inside the archive,
		// not where it happened to live on this server's disk. Naming them
		// by absolute path leaks the storage layout into every download and
		// re-extracts to a full path tree instead of the project.
		t.Run("names entries relative to the zipped root", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/zip_from_path"
			os.MkdirAll(pathPrefix+"/conf.d", 0755)
			os.WriteFile(pathPrefix+"/Dockerfile", []byte("FROM nginx"), 0644)
			os.WriteFile(pathPrefix+"/conf.d/default.conf", []byte("server {}"), 0644)

			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromPath(pathPrefix)
			if err != nil {
				t.Fatal(err)
			}

			zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
			if err != nil {
				t.Fatal(err)
			}

			got := []string{}
			for _, f := range zipReader.File {
				got = append(got, f.Name)
			}
			sort.Strings(got)

			want := []string{"Dockerfile", "conf.d/default.conf"}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("expected entries %v, but got %v", want, got)
			}
		})
	})
}

// unzipFiles builds a zip from files, in order, and extracts it into the
// shared unzip target directory, returning that directory alongside whatever
// UnzipToPath returned. Building the archive is setup, not the thing under
// test, so a failure there is fatal rather than a returned error.
func unzipFiles(t *testing.T, files []zipfile.File, unwrap bool) (string, error) {
	t.Helper()

	pathPrefix := PATH_PREFIX + "/unzip_to_path"
	os.MkdirAll(pathPrefix, 0755)

	s := zipservice.NewService(logger.NewMockLogger(), zipservice.Limits{})
	buf, err := s.ZipFromFiles(&files)
	if err != nil {
		t.Fatal(err)
	}

	return pathPrefix, s.UnzipToPath(buf, pathPrefix, unwrap)
}

// unzipFileMap is unzipFiles for cases where entry order is irrelevant.
func unzipFileMap(t *testing.T, fileMap map[string]string, unwrap bool) (string, error) {
	t.Helper()

	files := []zipfile.File{}
	for path, body := range fileMap {
		files = append(files, zipfile.File{Path: path, Body: body})
	}

	return unzipFiles(t, files, unwrap)
}

// assertFileBody fails unless path holds exactly body.
func assertFileBody(t *testing.T, path string, body string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != body {
		t.Errorf("expected %s at %s, but got %s", body, path, string(data))
	}
}

// assertNotExists fails unless path is absent.
func assertNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s NOT to exist, stat err: %v", path, err)
	}
}

func TestZipService_UnzipToPath(t *testing.T) {
	logger := logger.NewMockLogger()
	t.Run("path not exists", func(t *testing.T) {
		s := zipservice.NewService(logger, zipservice.Limits{})
		err := s.UnzipToPath(nil, PATH_PREFIX+"/not_exists", false)
		if err == nil {
			t.Error("expected error, but got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Run("single depth files", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			fileMap := map[string]string{
				"hello.txt": "hello",
				"world.txt": "world",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, false)
			if err != nil {
				t.Error(err)
			}

			for path, body := range fileMap {
				data, err := os.ReadFile(pathPrefix + "/" + path)
				if err != nil {
					t.Error(err)
				}

				if body != string(data) {
					t.Errorf("expected %s, but got %s", body, string(data))
				}
			}
		})

		t.Run("multi depth files", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			fileMap := map[string]string{
				"hi.txt":      "hi",
				"a/b/c/d.txt": "hello",
				"a/b/c/e.txt": "world",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, false)
			if err != nil {
				t.Error(err)
			}

			for path, body := range fileMap {
				data, err := os.ReadFile(pathPrefix + "/" + path)
				if err != nil {
					t.Error(err)
				}

				if body != string(data) {
					t.Errorf("expected %s, but got %s", body, string(data))
				}
			}
		})

		t.Run("unwrap if single directory", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			fileMap := map[string]string{
				"parent/hello.txt": "hello",
				"parent/world.txt": "world",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, true)
			if err != nil {
				t.Error(err)
			}

			for path, body := range fileMap {
				filename := filepath.Base(path)
				data, err := os.ReadFile(pathPrefix + "/" + filename)
				if err != nil {
					t.Fatal(err)
				}

				if body != string(data) {
					t.Errorf("expected %s, but got %s", body, string(data))
				}
			}
		})

		t.Run("unwrap if multi depth directory", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			fileMap := map[string]string{
				"parent/child/hello.txt": "hello",
				"parent/child/world.txt": "world",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, true)
			if err != nil {
				t.Error(err)
			}

			for path, body := range fileMap {
				filename := filepath.Base(path)
				data, err := os.ReadFile(pathPrefix + "/child/" + filename)
				if err != nil {
					t.Fatal(err)
				}

				if body != string(data) {
					t.Errorf("expected %s, but got %s", body, string(data))
				}
			}
		})

		t.Run("unzip to existing directory should unwrap single depth source", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			os.MkdirAll(pathPrefix+"/parent", 0755)
			os.Create(pathPrefix + "hello.txt")
			fileMap := map[string]string{
				"another/hello.txt": "hello",
				"another/world.txt": "world",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, true)
			if err != nil {
				t.Error(err)
			}

			for path, body := range fileMap {
				filename := filepath.Base(path)
				data, err := os.ReadFile(pathPrefix + "/" + filename)
				if err != nil {
					t.Fatal(err)
				}

				if body != string(data) {
					t.Errorf("expected %s, but got %s", body, string(data))
				}
			}
		})

		// Dotfiles are ordinary project files (.env, .htaccess, .dockerignore,
		// .github/) and must survive extraction. The blanket "skip anything
		// with a dot-prefixed path segment" rule this replaces silently
		// deleted them, which broke a contestant's project in production with
		// no error anywhere. Containment against zip-slip is isWithinDir's
		// job — see the "rejects a zip-slip entry" test below.
		t.Run("extracts dotfiles", func(t *testing.T) {
			setup()
			defer teardown()

			fileMap := map[string]string{
				".env":                     "SECRET=1",
				".gitignore":               "node_modules",
				".vscode/settings.json":    "{}",
				"src/.eslintrc":            "{}",
				".github/workflows/ci.yml": "on: push",
				"Dockerfile":               "FROM nginx",
			}

			pathPrefix, err := unzipFileMap(t, fileMap, false)
			if err != nil {
				t.Fatal(err)
			}

			for path, body := range fileMap {
				assertFileBody(t, pathPrefix+"/"+path, body)
			}
		})

		// The production shape: a contestant zips their project FOLDER, and
		// the folder contains dotfiles. Unwrap and dotfile handling have to
		// work together, not just individually.
		t.Run("unwraps a wrapper containing dotfiles", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"myproject/.env":       "SECRET=1",
				"myproject/.htaccess":  "RewriteEngine On",
				"myproject/Dockerfile": "FROM nginx",
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/.env", "SECRET=1")
			assertFileBody(t, pathPrefix+"/.htaccess", "RewriteEngine On")
			assertFileBody(t, pathPrefix+"/Dockerfile", "FROM nginx")
		})

		// A leading ".." is only an escape when it is the WHOLE segment.
		// "..config.txt" is a perfectly ordinary filename that the old
		// dot-prefix rule also ate.
		t.Run("extracts a file whose name starts with two dots", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"..config.txt": "two dots",
			}, false)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/..config.txt", "two dots")
		})

		t.Run("skips archiver junk", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"Dockerfile":                      "FROM nginx",
				"src/index.js":                    "code",
				".DS_Store":                       "junk",
				"src/.DS_Store":                   "junk",
				"._Dockerfile":                    "junk",
				"__MACOSX/myproject/._Dockerfile": "junk",
				"Thumbs.db":                       "junk",
				"desktop.ini":                     "junk",
			}, false)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/Dockerfile", "FROM nginx")
			assertFileBody(t, pathPrefix+"/src/index.js", "code")

			assertNotExists(t, pathPrefix+"/.DS_Store")
			assertNotExists(t, pathPrefix+"/src/.DS_Store")
			assertNotExists(t, pathPrefix+"/._Dockerfile")
			assertNotExists(t, pathPrefix+"/__MACOSX")
			assertNotExists(t, pathPrefix+"/Thumbs.db")
			assertNotExists(t, pathPrefix+"/desktop.ini")
		})

		// The guard rail for letting dotfiles through: macOS Finder's
		// "Compress" writes a sibling __MACOSX/ tree next to the real folder.
		// Those entries used to vanish because their basenames start with a
		// dot, so __MACOSX/ was never created and tmpdir held exactly one
		// entry — unwrap worked by accident. Extracting them instead would
		// give tmpdir two entries, silently disable unwrap, and bury the
		// project (and its Dockerfile) one level too deep for every Mac user.
		t.Run("a macOS-style archive still unwraps", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"myproject/Dockerfile":            "FROM nginx",
				"__MACOSX/myproject/._Dockerfile": "junk",
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/Dockerfile", "FROM nginx")
			assertNotExists(t, pathPrefix+"/myproject")
			assertNotExists(t, pathPrefix+"/__MACOSX")
		})

		t.Run("a root .DS_Store does not block unwrap", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"myproject/Dockerfile": "FROM nginx",
				".DS_Store":            "junk",
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/Dockerfile", "FROM nginx")
			assertNotExists(t, pathPrefix+"/myproject")
		})

		t.Run("do not unwrap if single root entry is file", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			fileMap := map[string]string{
				"README.txt": "hello",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, true)
			if err != nil {
				t.Error(err)
			}

			data, err := os.ReadFile(pathPrefix + "/README.txt")
			if err != nil {
				t.Fatal(err)
			}

			if string(data) != "hello" {
				t.Fatalf("expected hello, but got %s", string(data))
			}
		})

		t.Run("do not unwrap if root has multiple entries", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)
			fileMap := map[string]string{
				"alpha.txt":       "alpha",
				"nested/beta.txt": "beta",
			}
			files := &[]zipfile.File{}
			for path, body := range fileMap {
				(*files) = append(*files, zipfile.File{Path: path, Body: body})
			}
			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(files)
			if err != nil {
				t.Error(err)
			}

			err = s.UnzipToPath(buf, pathPrefix, true)
			if err != nil {
				t.Error(err)
			}

			alpha, err := os.ReadFile(pathPrefix + "/alpha.txt")
			if err != nil {
				t.Fatal(err)
			}
			if string(alpha) != "alpha" {
				t.Fatalf("expected alpha, but got %s", string(alpha))
			}

			beta, err := os.ReadFile(pathPrefix + "/nested/beta.txt")
			if err != nil {
				t.Fatal(err)
			}
			if string(beta) != "beta" {
				t.Fatalf("expected beta, but got %s", string(beta))
			}
		})

		// Regression tests for the unwrap collision hit in production. unwrap
		// used to os.Rename the wrapper's children up into tmpdir — the very
		// directory the wrapper itself still lived in. A child named exactly
		// like the wrapper therefore renamed onto the wrapper and the whole
		// submission died with "rename ...: file exists". Moving the copyDir
		// SOURCE instead of moving files makes the collision impossible: no
		// rename happens at all.
		t.Run("unwrap when a child directory shares the wrapper name", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"myproject/Dockerfile":           "FROM nginx",
				"myproject/myproject/index.html": "<h1>hi</h1>",
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/Dockerfile", "FROM nginx")
			assertFileBody(t, pathPrefix+"/myproject/index.html", "<h1>hi</h1>")
		})

		t.Run("unwrap when a child file shares the wrapper name", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"myproject/Dockerfile": "FROM nginx",
				"myproject/myproject":  "i am a file",
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/Dockerfile", "FROM nginx")
			assertFileBody(t, pathPrefix+"/myproject", "i am a file")
		})

		t.Run("unwrap when the wrapper contains only the same-named child", func(t *testing.T) {
			setup()
			defer teardown()

			pathPrefix, err := unzipFileMap(t, map[string]string{
				"myproject/myproject/index.html": "<h1>hi</h1>",
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			assertFileBody(t, pathPrefix+"/myproject/index.html", "<h1>hi</h1>")
		})
	})

	// "a" is extracted as a regular file first, so creating the directory
	// "a/" afterwards fails with ENOTDIR. That MkdirAll error used to be
	// discarded, letting the unzip report success while silently dropping
	// everything the directory should have contained — the same
	// fail-without-telling-anyone shape as the dotfile bug.
	t.Run("fails when a directory entry cannot be created", func(t *testing.T) {
		setup()
		defer teardown()

		_, err := unzipFiles(t, []zipfile.File{
			{Path: "a", Body: "i am a file"},
			{Path: "a/", Body: ""},
		}, false)

		if err == nil {
			t.Fatal("expected an error, but got nil")
		}
	})

	t.Run("limits", func(t *testing.T) {
		// unzipWithLimits extracts buf under the given limits.
		unzipWithLimits := func(t *testing.T, buf *bytes.Buffer, limits zipservice.Limits) (string, error) {
			t.Helper()

			pathPrefix := PATH_PREFIX + "/unzip_to_path"
			os.MkdirAll(pathPrefix, 0755)

			s := zipservice.NewService(logger, limits)
			return pathPrefix, s.UnzipToPath(buf, pathPrefix, false)
		}

		// zipOf builds an ordinary archive from files.
		zipOf := func(t *testing.T, files []zipfile.File) *bytes.Buffer {
			t.Helper()

			s := zipservice.NewService(logger, zipservice.Limits{})
			buf, err := s.ZipFromFiles(&files)
			if err != nil {
				t.Fatal(err)
			}
			return buf
		}

		t.Run("rejects an archive extracting past the byte limit", func(t *testing.T) {
			setup()
			defer teardown()

			buf := zipOf(t, []zipfile.File{
				{Path: "big.txt", Body: strings.Repeat("a", 4096)},
			})

			pathPrefix, err := unzipWithLimits(t, buf, zipservice.Limits{MaxBytes: 1024})
			if !errors.Is(err, zipservice.ErrorZipTooLarge) {
				t.Fatalf("expected ErrorZipTooLarge, got %v", err)
			}

			// The unzip failed as a whole, so nothing may have reached the
			// target directory.
			assertNotExists(t, pathPrefix+"/big.txt")
		})

		// The limit is about what lands on disk, so it is counted per
		// archive, not per entry: many small files must add up.
		t.Run("counts the byte limit across all entries", func(t *testing.T) {
			setup()
			defer teardown()

			buf := zipOf(t, []zipfile.File{
				{Path: "a.txt", Body: strings.Repeat("a", 600)},
				{Path: "b.txt", Body: strings.Repeat("b", 600)},
			})

			if _, err := unzipWithLimits(t, buf, zipservice.Limits{MaxBytes: 1024}); !errors.Is(err, zipservice.ErrorZipTooLarge) {
				t.Fatalf("expected ErrorZipTooLarge, got %v", err)
			}
		})

		t.Run("accepts an archive exactly at the byte limit", func(t *testing.T) {
			setup()
			defer teardown()

			body := strings.Repeat("a", 1024)
			buf := zipOf(t, []zipfile.File{{Path: "exact.txt", Body: body}})

			pathPrefix, err := unzipWithLimits(t, buf, zipservice.Limits{MaxBytes: 1024})
			if err != nil {
				t.Fatalf("expected the archive to be accepted, got %v", err)
			}

			assertFileBody(t, pathPrefix+"/exact.txt", body)
		})

		// The actual zip-bomb shape: an archive small enough to sail through
		// any upload-size check, expanding to something that fills the disk.
		// This is why the limit is on extracted bytes rather than on the
		// bytes the contestant uploaded.
		//
		// Note the limit deliberately counts bytes as they are written
		// instead of reading f.UncompressedSize64 up front. Go's archive/zip
		// makes the header trustworthy — it returns ErrFormat if the stream
		// does not decompress to exactly the declared size — so both would
		// be correct here. Counting as we go is what stops the copy MID
		// ENTRY, so a 100 GB bomb never reaches the disk in the first place.
		t.Run("rejects a small archive that expands past the byte limit", func(t *testing.T) {
			setup()
			defer teardown()

			// A megabyte of one repeated byte deflates to a couple of KiB.
			buf := zipOf(t, []zipfile.File{
				{Path: "bomb.txt", Body: strings.Repeat("a", 1024*1024)},
			})

			const limit = 64 * 1024
			if buf.Len() >= limit {
				t.Fatalf("expected the archive itself to be well under the limit to make the point, got %d bytes", buf.Len())
			}

			pathPrefix, err := unzipWithLimits(t, buf, zipservice.Limits{MaxBytes: limit})
			if !errors.Is(err, zipservice.ErrorZipTooLarge) {
				t.Fatalf("expected ErrorZipTooLarge, got %v", err)
			}

			assertNotExists(t, pathPrefix+"/bomb.txt")
		})

		t.Run("rejects an archive with more entries than the limit", func(t *testing.T) {
			setup()
			defer teardown()

			buf := zipOf(t, []zipfile.File{
				{Path: "a.txt", Body: "a"},
				{Path: "b.txt", Body: "b"},
				{Path: "c.txt", Body: "c"},
			})

			if _, err := unzipWithLimits(t, buf, zipservice.Limits{MaxEntries: 2}); !errors.Is(err, zipservice.ErrorZipTooManyEntries) {
				t.Fatalf("expected ErrorZipTooManyEntries, got %v", err)
			}
		})

		t.Run("accepts an archive exactly at the entry limit", func(t *testing.T) {
			setup()
			defer teardown()

			buf := zipOf(t, []zipfile.File{
				{Path: "a.txt", Body: "a"},
				{Path: "b.txt", Body: "b"},
			})

			if _, err := unzipWithLimits(t, buf, zipservice.Limits{MaxEntries: 2}); err != nil {
				t.Fatalf("expected the archive to be accepted, got %v", err)
			}
		})

		// The escape hatch: a non-positive limit turns that limit off, which
		// is what the zero value used by every other test in this file means.
		t.Run("treats a non-positive limit as unlimited", func(t *testing.T) {
			setup()
			defer teardown()

			body := strings.Repeat("a", 4096)
			buf := zipOf(t, []zipfile.File{{Path: "big.txt", Body: body}})

			pathPrefix, err := unzipWithLimits(t, buf, zipservice.Limits{MaxBytes: -1, MaxEntries: -1})
			if err != nil {
				t.Fatalf("expected no limit to apply, got %v", err)
			}

			assertFileBody(t, pathPrefix+"/big.txt", body)
		})
	})

	// Regression test for A3 (zip-slip): an entry whose name escapes the
	// extraction directory must fail the whole unzip with
	// ErrorZipEntryEscapesTarget — not be silently skipped. (Silent skipping
	// is what the old check order did: isHidden happens to match every
	// "../" segment, so the containment guard never executed and the archive
	// just quietly lost the entry.)
	t.Run("rejects a zip-slip entry", func(t *testing.T) {
		setup()
		defer teardown()

		pathPrefix := PATH_PREFIX + "/unzip_to_path"
		os.MkdirAll(pathPrefix, 0755)

		// A unique escape filename so the not-escaped assertion can never
		// collide with residue from other tests or runs. UnzipToPath
		// extracts into a fresh os.MkdirTemp("", "unzip") directory before
		// moving results into the target, so an actual escape would land
		// RELATIVE TO THE SYSTEM TEMP DIR (the single ".." cancels the
		// random unzipXXXX component, leaving TempDir()/evilName) — not
		// anywhere under PATH_PREFIX. A single "../" on purpose: it fully
		// exercises the containment guard while keeping the worst case of a
		// broken guard (e.g. during revert-and-confirm verification) INSIDE
		// os.TempDir, never at the filesystem root.
		evilName := fmt.Sprintf("evil-%d.txt", time.Now().UnixNano())
		escaped := filepath.Clean(filepath.Join(os.TempDir(), "unzip-any", "../"+evilName))
		defer os.Remove(escaped)

		files := &[]zipfile.File{
			{Path: "safe.txt", Body: "safe"},
			{Path: "../" + evilName, Body: "pwned"},
		}
		s := zipservice.NewService(logger, zipservice.Limits{})
		buf, err := s.ZipFromFiles(files)
		if err != nil {
			t.Fatal(err)
		}

		err = s.UnzipToPath(buf, pathPrefix, false)
		if !errors.Is(err, zipservice.ErrorZipEntryEscapesTarget) {
			t.Fatalf("expected ErrorZipEntryEscapesTarget, got %v", err)
		}

		if _, err := os.Stat(escaped); !os.IsNotExist(err) {
			t.Fatalf("expected %s NOT to exist outside the extraction directory, stat err: %v", escaped, err)
		}

		// The unzip failed as a whole, so nothing may have reached the
		// target directory — not even the benign entry.
		if _, err := os.Stat(pathPrefix + "/safe.txt"); !os.IsNotExist(err) {
			t.Fatalf("expected no partial extraction into the target directory, stat err: %v", err)
		}
	})
}
