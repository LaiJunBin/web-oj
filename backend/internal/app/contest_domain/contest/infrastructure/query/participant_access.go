package query

import (
	"gorm.io/gorm"
)

// isContestParticipant reports whether the user was added to the contest
// directly or through a group; soft-deleted groups must not grant access.
func isContestParticipant(db *gorm.DB, contestID, userID int32) (bool, error) {
	var count int64
	err := db.Raw(`
		SELECT COUNT(*) FROM contest_participants cp
		WHERE cp.contest_id = ?
		  AND (
		    cp.user_id = ?
		    OR EXISTS (
		      SELECT 1 FROM user_groups ug
		      INNER JOIN `+"`groups`"+` g ON g.id = ug.group_id AND g.deleted_at IS NULL
		      WHERE ug.group_id = cp.group_id AND ug.user_id = ?
		    )
		  )
	`, contestID, userID, userID).Scan(&count).Error
	return count > 0, err
}
