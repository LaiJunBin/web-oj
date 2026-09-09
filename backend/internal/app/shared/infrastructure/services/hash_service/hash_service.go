package hashservice

import (
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"

	"golang.org/x/crypto/bcrypt"
)

type HashService struct {
	logger sharedtypes.Logger
}

func NewHashService(logger sharedtypes.Logger) sharedservice.HashService {
	return &HashService{
		logger: logger,
	}
}

func (hs *HashService) Hash(text string) (string, error) {
	hs.logger.Debugf("Hashing text")
	hashed, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		hs.logger.Errorf("Failed to hash text: %v", err)
		return "", err
	}
	hs.logger.Debugf("Successfully hashed text")
	return string(hashed), nil
}

func (hs *HashService) Compare(text, hashed string) bool {
	hs.logger.Debugf("Comparing text with hashed value")
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(text))
	if err != nil {
		hs.logger.Warnf("Hash comparison failed: %v", err)
		return false
	}
	hs.logger.Debugf("Hash comparison succeeded")
	return true
}
