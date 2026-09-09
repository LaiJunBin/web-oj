package repository

type UserRepository interface {
	// FindUsernameByID resolves the username of the user who owns a
	// submission, including soft-deleted users: a submission already in the
	// pipeline must stay judgeable even if the account is removed
	// mid-contest. A soft-deleted row's stored username was renamed to free
	// the unique index; the implementation restores the ORIGINAL name, so
	// SUBMIT_USERNAME keeps addressing the per-seat resources it named.
	FindUsernameByID(userID int32) (string, error)
}
