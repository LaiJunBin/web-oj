package service

// LoginAttemptForgetter clears the login rate-limit record for an account.
//
// Login throttling lives in the transport layer (a middleware in the shared
// context), but a password reset here should let the contestant sign in with
// the new password immediately, not be turned away by the failed attempts that
// led staff to reset it. The need is expressed here as the narrowest interface
// that states it, and the middleware's limiter is bound to it at the
// composition root — the same bridge pattern as SessionRevoker.
type LoginAttemptForgetter interface {
	// ForgetUser drops the throttling record for username across all client IPs.
	// Safe to call for a user with no recorded attempts.
	ForgetUser(username string)
}
