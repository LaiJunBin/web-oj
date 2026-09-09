package submission

import (
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

type Submission struct {
	id         string
	userID     int32
	judgmentID string
	status     *Status
}

type SubmissionConfig struct {
	ID         string
	UserID     int32
	JudgmentID string
	Status     string
}

func NewSubmission(config *SubmissionConfig) (*Submission, error) {
	s, ok := StatusMap[config.Status]
	if !ok {
		return nil, httperror.New(http.StatusBadRequest, "invalid status")
	}

	return &Submission{
		id:         config.ID,
		userID:     config.UserID,
		judgmentID: config.JudgmentID,
		status:     s,
	}, nil
}

func (s *Submission) ID() string {
	return s.id
}

func (s *Submission) UserID() int32 {
	return s.userID
}

func (s *Submission) JudgmentID() string {
	return s.judgmentID
}

func (s *Submission) Status() *Status {
	return s.status
}

func (s *Submission) CanCancel() bool {
	return s.status.IsQueued() || s.status.IsJudging()
}
