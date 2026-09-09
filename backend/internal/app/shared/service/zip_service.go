package sharedservice

import (
	"bytes"
	"judge-backend/internal/app/shared/types/zipfile"
)

type ZipService interface {
	CheckZip(buf *bytes.Buffer) error
	// ReadZipToMap reads every file entry of an archive into memory, keyed by
	// entry name, under the same size/entry-count limits UnzipToPath enforces
	// on disk, and rejecting duplicate names. It is the bounded replacement for
	// "validate the zip parses, then io.ReadAll every entry into a map", which
	// let a small archive expand into gigabytes of resident memory.
	ReadZipToMap(buf *bytes.Buffer) (map[string][]byte, error)
	ZipFromFiles(files *[]zipfile.File) (*bytes.Buffer, error)
	ZipFromPath(path string) (*bytes.Buffer, error)
	UnzipToPath(zip *bytes.Buffer, path string, unwrap bool) error
}
