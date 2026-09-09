package sharedservice

import "bytes"

type FileSystemService interface {
	ListFileRecursive(path string) ([]string, error)
	Exists(path string) bool
	MkdirAll(path string) error
	Move(src, dst string) error
	RemoveAll(path string) error
	WriteFile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	TarPath(path string) (*bytes.Buffer, error)
}
