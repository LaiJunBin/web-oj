package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID   int `gorm:"primaryKey;autoIncrement"`
	Name string
	// Username is the login identity — uniqueness is enforced at the DB
	// level so a double-run of the seeders (or a duplicated roster import)
	// fails loudly instead of silently creating twin accounts. The index
	// has NO partial-index WHERE clause, so on every database (not only
	// MySQL) a soft-deleted row keeps occupying its username — which is
	// why deletion renames it (DeletedUsername) to free the name for
	// re-creation.
	Username string `gorm:"uniqueIndex;size:191"`
	Password string
	Roles    int
}

// DeletedUsernameMarker separates a soft-deleted row's original username from
// the row-id suffix that frees the name on the unqualified unique index (see
// the Username field comment). Same scheme as DeletedImageTagMarker. The
// identity domain duplicates the value (username.DeletedUsernameMarker, pinned
// equal by a test) because the domain layer does not import shared/model.
const DeletedUsernameMarker = "__deleted_"

// DeletedUsername renames username for its soft-deleted row. The row id keeps
// repeated delete/recreate cycles of the same name unique. An over-long
// username truncates the BASE to honour the column's size:191 — the username
// value object has no upper bound, so without this a ~190-rune name would
// turn the delete itself into a 1406 error. Byte-based truncation (matching
// DeletedImageTag) can split a multibyte rune; accepted, because real
// usernames are ASCII and the row is dead.
func DeletedUsername(username string, id int) string {
	suffix := fmt.Sprintf("%s%d", DeletedUsernameMarker, id)
	if len(username)+len(suffix) > 191 {
		username = username[:191-len(suffix)]
	}
	return username + suffix
}

// OriginalUsername strips DeletedUsername's suffix, recovering the name a
// soft-deleted row had while alive. The judge pipeline depends on this: an
// in-flight submission stays judgeable after its account is removed, and the
// SUBMIT_USERNAME it hands the judge containers must keep addressing the
// per-seat resources created under the ORIGINAL name.
//
// Only a trailing marker followed by pure digits is stripped, and only the
// LAST occurrence — so a pre-guard live user named "x__deleted_9" (creating
// one is rejected as reserved nowadays) who is then deleted round-trips
// correctly (one suffix added, one stripped). A
// base that DeletedUsername truncated cannot be fully recovered; accepted
// for the same reason the truncation is.
func OriginalUsername(stored string) string {
	idx := strings.LastIndex(stored, DeletedUsernameMarker)
	if idx < 0 {
		return stored
	}
	digits := stored[idx+len(DeletedUsernameMarker):]
	if digits == "" {
		return stored
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return stored
		}
	}
	return stored[:idx]
}
