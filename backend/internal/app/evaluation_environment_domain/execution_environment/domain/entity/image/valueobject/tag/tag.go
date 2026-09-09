package imagetag

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidTagIsEmpty = errors.New("tag is empty")
	ErrInvalidTagFormat  = errors.New("invalid tag format")
)

var tagRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]+:[a-zA-Z0-9-_.]+$`)

// DeletedTagMarker duplicates model.DeletedImageTagMarker — the domain layer
// does not import shared/model, so the value is pinned equal by a test
// instead. A tag containing it could collide with a soft-deleted row's
// renamed tag, making that row's delete fail on the unique index.
const DeletedTagMarker = "__deleted_"

// ContainsDeletedMarker reports whether a tag uses the rename-on-delete
// namespace. Deliberately NOT part of Validate(): soft-deleted rows
// legitimately hold marker tags, and build-history hydration
// (build_repository's imageIDToTag) reads them back through New — the rule
// is enforced only at the CREATION choke point (ImageFactory, bundle
// validate).
func ContainsDeletedMarker(value string) bool {
	return strings.Contains(value, DeletedTagMarker)
}

type Tag struct {
	value string
}

func New(value string) (Tag, error) {
	tag := Tag{value: value}
	if err := tag.Validate(); err != nil {
		return Tag{}, err
	}

	return tag, nil
}

func MustNew(value string) Tag {
	tag, err := New(value)
	if err != nil {
		panic(err)
	}

	return tag
}

func (t Tag) String() string {
	return t.value
}

func (t Tag) Validate() error {
	if t.value == "" {
		return ErrInvalidTagIsEmpty
	}

	if !tagRegex.MatchString(t.value) {
		return ErrInvalidTagFormat
	}

	return nil
}
