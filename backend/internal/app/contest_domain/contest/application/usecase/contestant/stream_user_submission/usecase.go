package streamusersubmission

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"
)

var (
	ErrorInvalidUserID       = httperror.New(http.StatusBadRequest, "invalid user id")
	ErrorInvalidSubmissionID = httperror.New(http.StatusBadRequest, "invalid submission id")
	ErrorSubmissionNotFound  = httperror.New(http.StatusNotFound, "submission not found")
)

type Input struct {
	UserID       int32
	SubmissionID string
}

// sampleSetTTL bounds how stale an open stream's sample-testcase set may get:
// testcase visibility CAN change mid-contest, and a set frozen at open kept
// streaming a testcase's output after it was flipped to hidden. Five seconds
// keeps the worst-case exposure short while capping the re-resolution cost at
// one query per stream per TTL, however dense the judge-update burst. A
// constant on purpose — this is not an operator knob.
const sampleSetTTL = 5 * time.Second

type Output struct {
	SubmissionID string
	// ResolveSampleTestcaseIDs returns the set of testcases whose run output
	// the stream may push right now, mirroring the REST detail's filtering.
	// Memoised with sampleSetTTL; the presenter calls it once per delivered
	// frame, on the stream's single writer goroutine (no locking needed).
	ResolveSampleTestcaseIDs func() map[int32]bool
}

// newSampleSetResolver memoises fetch behind ttl. A failed refresh keeps the
// last known set — never wider, never empty — and leaves the memo expired so
// the next event retries.
func newSampleSetResolver(
	initial []int32,
	fetch func() ([]int32, error),
	ttl time.Duration,
	now func() time.Time,
	logger sharedtypes.Logger,
) func() map[int32]bool {
	toSet := func(ids []int32) map[int32]bool {
		set := make(map[int32]bool, len(ids))
		for _, id := range ids {
			set[id] = true
		}
		return set
	}

	cached := toSet(initial)
	fetchedAt := now()
	return func() map[int32]bool {
		if now().Sub(fetchedAt) <= ttl {
			return cached
		}
		ids, err := fetch()
		if err != nil {
			logger.Warnf("Failed to refresh the sample testcase set, keeping the previous one", "error", err)
			return cached
		}
		cached = toSet(ids)
		fetchedAt = now()
		return cached
	}
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.SubmissionQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.SubmissionQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.UserID <= 0 {
		return ErrorInvalidUserID
	}

	if input.SubmissionID == "" {
		return ErrorInvalidSubmissionID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to stream user submission", "userId", input.UserID, "submissionId", input.SubmissionID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.UserID, "submissionId", input.SubmissionID, "error", err)
		return nil, err
	}

	submission, err := u.query.GetUserSubmissionByID(input.UserID, input.SubmissionID)
	if err != nil {
		u.logger.Errorf("Failed to get user submission", "userId", input.UserID, "submissionId", input.SubmissionID, "error", err)
		return nil, err
	}

	if submission == nil {
		u.logger.Errorf("Submission not found", "userId", input.UserID, "submissionId", input.SubmissionID)
		return nil, ErrorSubmissionNotFound
	}

	u.logger.Infof("Successfully started streaming user submission", "userId", input.UserID, "submissionId", input.SubmissionID)
	contestProjectID := submission.ContestProjectID
	return &Output{
		SubmissionID: input.SubmissionID,
		ResolveSampleTestcaseIDs: newSampleSetResolver(
			submission.SampleTestcaseIDs,
			func() ([]int32, error) { return u.query.GetSampleTestcaseIDs(contestProjectID) },
			sampleSetTTL,
			time.Now,
			u.logger,
		),
	}, nil
}
