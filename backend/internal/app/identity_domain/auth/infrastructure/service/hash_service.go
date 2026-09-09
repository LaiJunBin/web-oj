package service

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type HashService struct {
	hashService sharedservice.HashService
	logger      sharedtypes.Logger
}

func NewHashService(
	hashService sharedservice.HashService,
	logger sharedtypes.Logger,
) service.HashService {
	return &HashService{
		hashService: hashService,
		logger:      logger,
	}
}

func (hs *HashService) Hash(text string) (string, error) {
	hs.logger.Debugf("Hashing text")
	hashed, err := hs.hashService.Hash(text)
	if err != nil {
		hs.logger.Errorf("Failed to hash text: %v", err)
		return "", err
	}
	hs.logger.Debugf("Text hashed successfully")
	return hashed, nil
}

func (hs *HashService) Compare(text, hashed string) bool {
	hs.logger.Debugf("Comparing text with hashed value")
	result := hs.hashService.Compare(text, hashed)
	if result {
		hs.logger.Debugf("Comparison successful")
	} else {
		hs.logger.Warnf("Comparison failed")
	}
	return result
}
