package service

import (
	"bytes"
	"judge-backend/internal/app/shared/types/submissiondynamicfield"
)

type SubmissionService interface {
	SaveProject(submissionID string, project *bytes.Buffer) error
	SaveDynamicFields(submissionID string, dynamicFields []*submissiondynamicfield.SubmissionDynamicField) error
	CleanupSubmissionFiles(submissionID string)
	// LoadMetadata reads back the dynamic-field metadata previously written by
	// SaveDynamicFields for a submission.
	LoadMetadata(submissionID string) (map[string]string, error)
}
