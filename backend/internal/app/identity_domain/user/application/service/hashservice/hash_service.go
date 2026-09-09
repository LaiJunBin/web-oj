package hashservice

type HashService interface {
	Hash(text string) (string, error)
	Compare(text, hashed string) bool
}
