package service

// SessionRevoker drops every session an account currently holds.
//
// Sessions belong to the auth bounded context, which owns the token store and
// the token lifetimes; the user context only knows that changing somebody's
// password should not leave whoever had the old one still signed in. So the
// need is expressed here as the narrowest interface that states it, and the
// auth repository is bound to it at the composition root — the same bridge
// pattern app/provider uses to keep the other contexts from importing each
// other.
type SessionRevoker interface {
	// RevokeAllSessions revokes every session held by the given user. It is
	// safe to call for a user with no sessions.
	RevokeAllSessions(userID int) error
}
