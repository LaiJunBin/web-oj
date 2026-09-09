package query

type ContestProjectQuery interface {
	CheckProjectAndGradingItemsExists(contestProjectID int32, gradingItemIDs []int32) error
	CheckUserIsParticipant(contestProjectID int32, userID int32) error
	CheckUserCanAccessProject(contestProjectID int32, userID int32) error
	// GetSubmissionCooldownSeconds returns this contest project's own
	// cooldown, or nil to fall back to the system default.
	GetSubmissionCooldownSeconds(contestProjectID int32) (*int32, error)
}
