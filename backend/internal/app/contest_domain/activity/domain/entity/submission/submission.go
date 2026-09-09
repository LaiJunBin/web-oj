package submission

import "time"

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
	ID               string
	UserID           int32
	ContestProjectID int32
	GradingItemIds   []int32
	IsPartial        bool
	Ignored          bool
	CreatedAt        time.Time
}

func New(config *SubmissionConfig) *Submission {
	return &Submission{
		id:               config.ID,
		userID:           config.UserID,
		contestProjectID: config.ContestProjectID,
		gradingItemIds:   config.GradingItemIds,
		isPartial:        config.IsPartial,
		ignored:          config.Ignored,
		createdAt:        config.CreatedAt,
	}
}

func (s *Submission) ID() string {
	return s.id
}

func (s *Submission) SetID(id string) {
	s.id = id
}

func (s *Submission) UserID() int32 {
	return s.userID
}

func (s *Submission) ContestProjectID() int32 {
	return s.contestProjectID
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

func (s *Submission) MarkAsIgnored() {
	s.ignored = true
}

func (s *Submission) MarkAsUnignored() {
	s.ignored = false
}
