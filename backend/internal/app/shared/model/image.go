package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Image struct {
	gorm.Model
	ID int `gorm:"primaryKey;autoIncrement"`
	// The partial-index WHERE clause only exists on databases that support
	// it (sqlite): MySQL has no partial indexes, so there the index is
	// simply UNIQUE(tag) and soft-deleted rows keep occupying their tag —
	// which is why deletion renames the tag (DeletedImageTag) instead of
	// relying on deleted_at falling out of the index.
	Tag         string `gorm:"size:255;uniqueIndex:tag,WHERE:deleted_at IS NULL"`
	MountPath   string
	Description string
	IsReady     bool
}

// DeletedImageTagMarker separates a soft-deleted row's original tag from the
// row-id suffix that frees the tag on MySQL (see the Tag field comment).
const DeletedImageTagMarker = "__deleted_"

// DeletedImageTag renames tag for its soft-deleted row. The suffix uses only
// tag-legal characters ([a-zA-Z0-9-_.], no second colon), so a renamed tag
// still passes imagetag validation when build history reads it back
// (imageIDToTag is unscoped); the row id makes repeated delete/recreate
// cycles of the same tag unique. An over-long tag truncates the BASE to
// honour the column's size:255 — in the pathological case where that cuts
// off the colon, build-history hydration of that dead row would fail
// validation, accepted because real tags are nowhere near 255 runes.
func DeletedImageTag(tag string, id int) string {
	suffix := fmt.Sprintf("%s%d", DeletedImageTagMarker, id)
	if len(tag)+len(suffix) > 255 {
		tag = tag[:255-len(suffix)]
	}
	return tag + suffix
}
