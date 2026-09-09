package hashservice

import (
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/hashservice"
)

type HashService struct {
	hashService sharedservice.HashService
	logger      sharedtypes.Logger
}

func NewHashService(
	hashService sharedservice.HashService,
	logger sharedtypes.Logger,
) hashservice.HashService {
	return &HashService{
		hashService: hashService,
		logger:      logger,
	}
}

func (hs *HashService) Hash(text string) (string, error) {
	hs.logger.Infof("Starting to hash text")

	result, err := hs.hashService.Hash(text)
	if err != nil {
		hs.logger.Errorf("Error hashing text", "error", err)
		return "", err
	}

	hs.logger.Infof("Successfully hashed text")
	return result, nil
}

func (hs *HashService) Compare(text, hashed string) bool {
	hs.logger.Infof("Starting to compare text with hash")

	result := hs.hashService.Compare(text, hashed)

	hs.logger.Infof("Successfully compared text with hash", "match", result)
	return result
}
