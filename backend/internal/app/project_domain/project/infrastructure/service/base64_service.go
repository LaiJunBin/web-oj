package service

import (
	"encoding/base64"
	b64 "judge-backend/internal/app/project_domain/project/application/service/base64"
)

type Base64Service struct {
}

func NewBase64Service() b64.Base64Service {
	return &Base64Service{}
}

func (s *Base64Service) Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func (s *Base64Service) Decode(base64Encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
