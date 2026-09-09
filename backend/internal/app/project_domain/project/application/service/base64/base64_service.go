package base64

type Base64Service interface {
	Encode(data string) string
	Decode(base64Encoded string) (string, error)
}
