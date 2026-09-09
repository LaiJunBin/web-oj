package dynamicfield

import (
	"errors"
	"strings"
)

var (
	ErrorInvalidName = errors.New("name is invalid, must not be empty")
	// ErrorUnsafeName rejects a field name that could escape the submission
	// files directory. A file-type field's name becomes a filename under
	// storage/submissions/<id>/files/, so a name like "../../config.yaml" would
	// let a contestant's submitted content be written outside that directory.
	// The final write site also contains the path, but a malicious project
	// bundle should be refused at import rather than relied on to fail safely
	// three layers later.
	ErrorUnsafeName = errors.New("name is invalid, must not contain path separators or traversal segments")
)

type Name struct {
	value string
}

func NewName(value string) (Name, error) {
	name := Name{value: value}
	if err := name.validate(); err != nil {
		return Name{}, err
	}

	return name, nil
}

func MustNewName(value string) Name {
	name, err := NewName(value)
	if err != nil {
		panic(err)
	}

	return name
}

func (n Name) validate() error {
	if n.value == "" {
		return ErrorInvalidName
	}

	// Block path traversal at the source. A separator or a "." / ".." segment is
	// never legitimate in a field key, and refusing them here keeps a hostile
	// bundle from ever reaching the filesystem. Charset is deliberately left
	// otherwise open so existing non-ASCII keys keep working.
	if strings.ContainsAny(n.value, `/\`) || strings.ContainsRune(n.value, 0) {
		return ErrorUnsafeName
	}
	if n.value == "." || n.value == ".." {
		return ErrorUnsafeName
	}

	return nil
}

func (n Name) String() string {
	return n.value
}

func (n Name) Equals(other Name) bool {
	return n.value == other.value
}
