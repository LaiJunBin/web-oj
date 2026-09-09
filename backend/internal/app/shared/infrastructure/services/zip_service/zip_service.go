package zipservice

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"

	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/zipfile"
	"os"
	"path/filepath"
	"strings"
)

// ErrorZipEntryEscapesTarget is returned by UnzipToPath when a zip entry's
// name (e.g. one starting with "../") would extract outside the target
// directory (a "zip-slip" attempt).
var ErrorZipEntryEscapesTarget = errors.New("zip entry escapes target directory")

var (
	// ErrorZipTooLarge is returned by UnzipToPath when an archive extracts
	// to more bytes than Limits.MaxBytes allows.
	ErrorZipTooLarge = errors.New("zip archive extracts to more than the allowed size")
	// ErrorZipTooManyEntries is returned by UnzipToPath when an archive
	// holds more entries than Limits.MaxEntries allows.
	ErrorZipTooManyEntries = errors.New("zip archive contains more entries than allowed")
	// ErrorZipDuplicateEntry is returned by ReadZipToMap when two entries share
	// a name, which would otherwise let a later entry silently overwrite an
	// earlier one in the resulting map.
	ErrorZipDuplicateEntry = errors.New("zip archive contains duplicate entry names")
)

// Limits bounds what a single archive is allowed to cost. A non-positive
// field disables that limit, so the zero value is "no limits" — convenient
// in tests, and the reason production wiring goes through NewZipLimits
// rather than constructing this inline.
type Limits struct {
	// MaxBytes caps the total EXTRACTED size across the whole archive. It
	// deliberately bounds what lands on disk rather than what was uploaded:
	// a zip bomb is small enough to pass any upload check and expands to
	// something that fills the volume.
	//
	// It is counted as bytes are written rather than read from
	// f.UncompressedSize64. Go's archive/zip does make the header
	// trustworthy — it fails the read unless the stream decompresses to
	// exactly the declared size — so either would be correct. Counting as we
	// go is what aborts a bomb MID ENTRY instead of after it has already
	// been written out.
	MaxBytes int64
	// MaxEntries caps how many entries the archive may contain, checked
	// before any of them is extracted.
	MaxEntries int
}

type Service struct {
	logger sharedtypes.Logger
	limits Limits
}

func NewService(logger sharedtypes.Logger, limits Limits) sharedservice.ZipService {
	return &Service{
		logger: logger,
		limits: limits,
	}
}

func (z *Service) CheckZip(buf *bytes.Buffer) error {
	z.logger.Debugf("Checking if buffer is a valid zip archive")
	_, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		z.logger.Warnf("Invalid zip archive: %v", err)
		return err
	}
	z.logger.Debugf("Buffer is a valid zip archive")
	return nil
}

// ReadZipToMap reads every file entry into memory under the archive's limits.
// It mirrors UnzipToPath's guards — entry count checked before reading anything,
// a single byte budget carried across all entries so the archive as a whole
// cannot exceed MaxBytes, and abort MID ENTRY the moment it does — but keeps the
// bytes in a map instead of writing them to disk. Duplicate names are rejected
// rather than silently collapsed.
func (z *Service) ReadZipToMap(buf *bytes.Buffer) (map[string][]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		z.logger.Warnf("Invalid zip archive: %v", err)
		return nil, err
	}

	if z.limits.MaxEntries > 0 && len(r.File) > z.limits.MaxEntries {
		z.logger.Errorf("Rejected archive with %d entries, limit is %d", len(r.File), z.limits.MaxEntries)
		return nil, fmt.Errorf("%w: %d entries, limit is %d", ErrorZipTooManyEntries, len(r.File), z.limits.MaxEntries)
	}

	// MaxInt64-1 rather than MaxInt64 so the budget+1 read below cannot overflow,
	// matching UnzipToPath.
	remaining := int64(math.MaxInt64 - 1)
	if z.limits.MaxBytes > 0 {
		remaining = z.limits.MaxBytes
	}

	files := make(map[string][]byte, len(r.File))
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if _, dup := files[f.Name]; dup {
			return nil, fmt.Errorf("%w: %q", ErrorZipDuplicateEntry, f.Name)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		var out bytes.Buffer
		// One byte past the budget is enough to prove the archive exceeds it and
		// stops here instead of reading the rest of a bomb into memory.
		written, err := io.CopyN(&out, rc, remaining+1)
		rc.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if written > remaining {
			return nil, ErrorZipTooLarge
		}
		remaining -= written
		files[f.Name] = out.Bytes()
	}

	return files, nil
}

func (z *Service) ZipFromFiles(files *[]zipfile.File) (*bytes.Buffer, error) {
	z.logger.Debugf("Creating zip archive from files")
	if files == nil {
		files = &[]zipfile.File{}
	}

	buf := new(bytes.Buffer)

	w := zip.NewWriter(buf)

	for _, file := range *files {
		f, err := w.Create(file.Path)
		if err != nil {
			z.logger.Errorf("Failed to create zip archive from files: %v", err)
			return nil, err
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			z.logger.Errorf("Failed to create zip archive from files: %v", err)
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		z.logger.Errorf("Failed to create zip archive from files: %v", err)
		return nil, err
	}

	z.logger.Debugf("Successfully created zip archive from files")
	return buf, nil
}

func (z *Service) pathExists(path string) bool {
	z.logger.Debugf("Checking if path exists: %s", path)
	_, err := os.Stat(path)
	exists := err == nil
	if exists {
		z.logger.Debugf("Path exists: %s", path)
	} else {
		z.logger.Warnf("Path does not exist: %s", path)
	}
	return exists
}

func (z *Service) ZipFromPath(path string) (*bytes.Buffer, error) {
	z.logger.Debugf("Creating zip archive from path: %s", path)
	if !z.pathExists(path) {
		z.logger.Warnf("Path does not exist: %s", path)
		return nil, os.ErrNotExist
	}

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
		z.logger.Errorf("Failed to create zip archive from path %s: %v", path, err)
		return nil, err
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		// Entry names describe where a file sits inside the archive, not
		// where it lives on this server. Naming them by absolute path would
		// leak the storage layout into every download and re-extract as a
		// full path tree rather than the project itself.
		name, err := filepath.Rel(path, file)
		if err != nil {
			z.logger.Errorf("Failed to create zip archive from path %s: %v", path, err)
			return nil, err
		}

		f, err := w.Create(filepath.ToSlash(name))
		if err != nil {
			z.logger.Errorf("Failed to create zip archive from path %s: %v", path, err)
			return nil, err
		}

		body, err := os.ReadFile(file)
		if err != nil {
			z.logger.Errorf("Failed to create zip archive from path %s: %v", path, err)
			return nil, err
		}

		_, err = f.Write(body)
		if err != nil {
			z.logger.Errorf("Failed to create zip archive from path %s: %v", path, err)
			return nil, err
		}
	}

	err = w.Close()
	if err != nil {
		z.logger.Errorf("Failed to create zip archive from path %s: %v", path, err)
		return nil, err
	}

	z.logger.Debugf("Successfully created zip archive from path: %s", path)
	return buf, nil
}

// isWithinDir reports whether path is dir itself or a descendant of it.
func isWithinDir(dir string, path string) bool {
	cleanDir := filepath.Clean(dir)
	return path == cleanDir || strings.HasPrefix(path, cleanDir+string(filepath.Separator))
}

// isArchiverJunk reports whether name is metadata an archiving tool invented
// rather than a file anyone authored. It deliberately does NOT match
// dotfiles at large: .env, .htaccess, .dockerignore and .github/ are real
// project files, and skipping them silently broke contestant projects.
// Extraction stays safe without a dot rule because isWithinDir rejects
// escaping entries outright, before this is consulted.
//
// The basename is taken by hand rather than with path.Base: zip entry names
// are always slash-separated so the logic is trivial, and "path" as an
// identifier is a parameter name throughout this file — importing the
// package of the same name would be shadowed wherever it is most tempting
// to use.
func isArchiverJunk(name string) bool {
	// macOS Finder writes a __MACOSX sidecar tree beside the real folder.
	// Only at the archive root: a nested directory a contestant happened to
	// name __MACOSX is their own file. The separator keeps "__MACOSXfoo/"
	// from matching.
	if strings.HasPrefix(name, "__MACOSX/") {
		return true
	}

	base := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		base = name[i+1:]
	}

	// AppleDouble resource forks accompanying each file.
	if strings.HasPrefix(base, "._") {
		return true
	}

	switch base {
	case ".DS_Store", "Thumbs.db", "desktop.ini":
		return true
	}

	return false
}

// extractEntry writes a single zip entry to fpath, spending at most budget
// bytes and reporting how many it used so the caller can carry the remainder
// to the next entry.
//
// It owns the entry reader for the whole call, so no error path can skip
// closing it, and every directory it creates is checked — a failed MkdirAll
// used to be discarded, letting an unzip report success while silently
// dropping the entries that directory should have held.
func (z *Service) extractEntry(f *zip.File, fpath string, budget int64) (int64, error) {
	if f.FileInfo().IsDir() {
		return 0, os.MkdirAll(fpath, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return 0, err
	}

	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	dst, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return 0, err
	}

	// Copy at most one byte past the budget: that is enough to prove the
	// archive exceeds it, and stops here rather than writing out the rest of
	// a bomb. CopyN reports io.EOF when the entry ended first, which is the
	// ordinary case and not an error.
	written, err := io.CopyN(dst, rc, budget+1)
	if err != nil && !errors.Is(err, io.EOF) {
		dst.Close()
		return written, err
	}

	if written > budget {
		dst.Close()
		return written, ErrorZipTooLarge
	}

	return written, dst.Close()
}

func (z *Service) UnzipToPath(zipBuf *bytes.Buffer, path string, unwrap bool) error {
	z.logger.Debugf("Unzipping archive to path: %s", path)
	if !z.pathExists(path) {
		z.logger.Warnf("Path does not exist: %s", path)
		return os.ErrNotExist
	}

	tmpdir, err := os.MkdirTemp("", "unzip")
	if err != nil {
		z.logger.Errorf("Failed to unzip archive to path %s: %v", path, err)
		return err
	}

	defer os.RemoveAll(tmpdir)

	r, err := zip.NewReader(bytes.NewReader(zipBuf.Bytes()), int64(zipBuf.Len()))
	if err != nil {
		z.logger.Errorf("Failed to unzip archive to path %s: %v", path, err)
		return err
	}

	// Checked before extracting anything, so an archive stuffed with entries
	// costs one comparison rather than a walk.
	if z.limits.MaxEntries > 0 && len(r.File) > z.limits.MaxEntries {
		z.logger.Errorf("Rejected archive with %d entries, limit is %d", len(r.File), z.limits.MaxEntries)
		return fmt.Errorf("%w: %d entries, limit is %d", ErrorZipTooManyEntries, len(r.File), z.limits.MaxEntries)
	}

	// remaining is the extraction budget carried across entries so the limit
	// applies to the archive as a whole. When disabled it starts effectively
	// unbounded, which keeps the accounting below branch-free — MaxInt64-1
	// rather than MaxInt64 so extractEntry's budget+1 cannot overflow.
	remaining := int64(math.MaxInt64 - 1)
	if z.limits.MaxBytes > 0 {
		remaining = z.limits.MaxBytes
	}

	for _, f := range r.File {
		// Containment first, and on its own: an escaping entry is an attack
		// and must fail the whole unzip loudly rather than be quietly
		// dropped. The skip below is a convenience filter for archiver
		// noise, never a security control.
		fpath := filepath.Join(tmpdir, f.Name)
		if !isWithinDir(tmpdir, fpath) {
			z.logger.Errorf("Rejected zip entry escaping target directory: %s", f.Name)
			return fmt.Errorf("%w: %s", ErrorZipEntryEscapesTarget, f.Name)
		}

		if isArchiverJunk(f.Name) {
			z.logger.Debugf("Skipping archiver metadata: %s", f.Name)
			continue
		}

		written, err := z.extractEntry(f, fpath, remaining)
		if err != nil {
			z.logger.Errorf("Failed to unzip archive to path %s: %v", path, err)
			return err
		}
		remaining -= written
	}

	// Unwrapping moves the COPY SOURCE down one level rather than moving the
	// wrapper's children up into tmpdir. Moving files used to mean renaming
	// each child into the directory the wrapper itself still occupied, so a
	// child named exactly like the wrapper renamed onto the wrapper and the
	// whole extraction died with "rename ...: file exists" — a real
	// production failure for a contestant whose project contained a folder
	// named after the project. Pointing copyDir at the wrapper achieves the
	// same result with no rename, so that collision cannot happen.
	srcDir := tmpdir
	if unwrap {
		entries, err := os.ReadDir(tmpdir)
		if err != nil {
			z.logger.Errorf("Failed to unzip archive to path %s: %v", path, err)
			return err
		}

		if len(entries) == 1 && entries[0].IsDir() {
			srcDir = filepath.Join(tmpdir, entries[0].Name())
		}
	}

	err = z.copyDir(srcDir, path)
	if err != nil {
		z.logger.Errorf("Failed to unzip archive to path %s: %v", path, err)
		return err
	}

	z.logger.Debugf("Successfully unzipped archive to path: %s", path)
	return nil
}

func (z *Service) copyDir(src string, dest string) error {
	z.logger.Debugf("Copying directory from %s to %s", src, dest)
	files, err := os.ReadDir(src)
	if err != nil {
		z.logger.Errorf("Failed to read directory %s: %v", src, err)
		return err
	}

	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		z.logger.Errorf("Failed to create directory %s: %v", dest, err)
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			dirName := file.Name()
			if err := z.copyDir(filepath.Join(src, dirName), filepath.Join(dest, dirName)); err != nil {
				return err
			}
		} else {
			fileName := file.Name()
			if err := z.CopyFile(filepath.Join(src, fileName), filepath.Join(dest, fileName)); err != nil {
				return err
			}
		}
	}

	z.logger.Debugf("Successfully copied directory from %s to %s", src, dest)
	return nil
}

func (z *Service) CopyFile(src string, dest string) error {
	z.logger.Debugf("Copying file from %s to %s", src, dest)
	srcFile, err := os.Open(src)
	if err != nil {
		z.logger.Errorf("Failed to open source file %s: %v", src, err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dest)
	if err != nil {
		z.logger.Errorf("Failed to create destination file %s: %v", dest, err)
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		z.logger.Errorf("Failed to copy file from %s to %s: %v", src, dest, err)
		return err
	}

	z.logger.Debugf("Successfully copied file from %s to %s", src, dest)
	return nil
}
