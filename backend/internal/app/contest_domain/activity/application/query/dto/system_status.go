package dto

import "time"

// SubmissionRatePoint is one minute-long bucket of the submission rate.
type SubmissionRatePoint struct {
	// Minute is the start of the bucket, truncated to the minute in UTC.
	Minute time.Time
	Count  int32
}

// JudgmentError is one judgment that ended in Error, reduced to what an admin
// needs in order to jump to the submission that produced it.
type JudgmentError struct {
	JudgmentID       string
	SubmissionID     string
	ContestProjectID int32

	// ErrorMessage is nil for judgments recorded before the failure reason was
	// captured — the field is optional on the stored document.
	ErrorMessage *string

	CreatedAt time.Time
}
