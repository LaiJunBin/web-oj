package service

type EncryptService interface {
	Encrypt(value string) (string, error)
	Decrypt(value string) (string, error)
}
