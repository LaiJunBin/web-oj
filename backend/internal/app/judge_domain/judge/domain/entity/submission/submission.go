package submission

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Submission struct {
	id               string
	userID           int32
	contestProjectID int32
	gradingItemIds   []int32
	isPartial        bool
	ignored          bool
	createdAt        time.Time
}

type SubmissionConfig struct {
	UserID           int32
	ContestProjectID int32
	GradingItemIds   []int32
	IsPartial        bool
	Ignored          bool
	CreatedAt        time.Time
}

// newID generates the submission's identity: 12 random bytes hex-encoded
// (24 chars, format-compatible with the historical Mongo ObjectID hex IDs).
// Uses only the standard library — the domain layer stays dependency-free.
func newID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate submission ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// New creates a submission with its identity already assigned. Anything that
// happens between creation and persistence (e.g. writing files under
// storage/submissions/{id}/) can rely on the ID being present — repositories
// reconstructing stored submissions overwrite it via SetID.
func New(config *SubmissionConfig) (*Submission, error) {
	if config.GradingItemIds == nil {
		config.GradingItemIds = []int32{}
	}

	createdAt := time.Now()
	if !config.CreatedAt.IsZero() {
		createdAt = config.CreatedAt
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}

	return &Submission{
		id:               id,
		userID:           config.UserID,
		contestProjectID: config.ContestProjectID,
		gradingItemIds:   config.GradingItemIds,
		isPartial:        config.IsPartial,
		ignored:          config.Ignored,
		createdAt:        createdAt,
	}, nil
}

func (s *Submission) ID() string {
	return s.id
}

func (s *Submission) ContestProjectID() int32 {
	return s.contestProjectID
}

func (s *Submission) UserID() int32 {
	return s.userID
}

func (s *Submission) GradingItemIds() []int32 {
	return s.gradingItemIds
}

func (s *Submission) IsPartial() bool {
	return s.isPartial
}

func (s *Submission) IsIgnored() bool {
	return s.ignored
}

func (s *Submission) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Submission) SetID(id string) {
	s.id = id
}
