package service

import (
	domainservice "judge-backend/internal/app/project_domain/project/domain/service"
)

type ValidParserNames []string

type ResultParserValidatorService struct {
	valid map[string]struct{}
}

func NewResultParserValidatorService(names ValidParserNames) domainservice.ResultParserValidator {
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[n] = struct{}{}
	}
	return &ResultParserValidatorService{valid: m}
}

func (s *ResultParserValidatorService) IsValid(parser string) bool {
	_, ok := s.valid[parser]
	return ok
}
