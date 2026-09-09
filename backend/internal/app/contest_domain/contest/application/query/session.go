package query

type SessionQuery interface {
	GetTTL(sessionID string) int
}
