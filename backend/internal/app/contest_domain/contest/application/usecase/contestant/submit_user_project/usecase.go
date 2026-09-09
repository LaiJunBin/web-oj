package submituserproject

import (
	"bytes"
	"errors"
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/contest_domain/contest/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/config"
	"net/http"
	"time"
)

var (
	ErrorInvalidData             = httperror.New(http.StatusBadRequest, "invalid data")
	ErrorInvalidUserID           = httperror.New(http.StatusBadRequest, "invalid user id")
	ErrorInvalidContestProjectID = httperror.New(http.StatusBadRequest, "invalid contest project id")
	ErrorProjectRequired         = httperror.New(http.StatusBadRequest, "project required")
	ErrorContestProjectNotFound  = httperror.New(http.StatusNotFound, "contest project not found")
	ErrorGradingItemNotFound     = httperror.New(http.StatusNotFound, "grading item not found")
	ErrorSubmissionNotStarted    = httperror.New(http.StatusBadRequest, "submission period has not started")
	ErrorSubmissionEnded         = httperror.New(http.StatusBadRequest, "submission period has ended")
)

// SubmissionCooldownError carries how much longer the contestant has to wait
// so the client can count down, rather than making it parse a message.
type SubmissionCooldownError struct {
	RetryAfterSeconds int32
}

func (e *SubmissionCooldownError) Error() string {
	return "submission cooldown has not elapsed"
}

type Input struct {
	UserID           int32
	ContestProjectID int32
	Project          *bytes.Buffer
	GradingItemIDs   []int32
	DynamicFields    map[string]string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query       query.ContestProjectQuery
	projectRepo repository.ProjectRepository
	cooldown    service.SubmissionCooldown
	publisher   service.ProjectSubmitPublisher
	config      *config.Config
	logger      sharedtypes.Logger
}

func NewUsecase(
	query query.ContestProjectQuery,
	projectRepo repository.ProjectRepository,
	cooldown service.SubmissionCooldown,
	publisher service.ProjectSubmitPublisher,
	config *config.Config,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:       query,
		projectRepo: projectRepo,
		cooldown:    cooldown,
		publisher:   publisher,
		config:      config,
		logger:      logger,
	}
}

// checkSubmissionWindow is the server-side eligibility gate: reads only ever
// showed the contest project as open, but nothing on the write side enforced
// it, so a direct API call could submit outside the window. FindByID already
// refuses a non-activated contest project (contest and project activation
// periods both) with a 404; that is re-wrapped so the response is
// indistinguishable from the contest project not existing at all. The window
// boundaries deliberately mirror the read side's model.ContestProject
// semantics — strictly after StartTime, up to and including EndTime — on the
// server clock.
func (u *usecase) checkSubmissionWindow(input *Input) error {
	p, err := u.projectRepo.FindByID(input.ContestProjectID)
	if err != nil {
		var he *httperror.Error
		if errors.As(err, &he) && he.Code == http.StatusNotFound {
			return ErrorContestProjectNotFound
		}
		return err
	}

	now := time.Now()
	if p.IsSubmissionNotStarted(now) {
		return ErrorSubmissionNotStarted
	}
	if p.IsSubmissionEnded(now) {
		return ErrorSubmissionEnded
	}
	return nil
}

// cooldownFor resolves the wait this contest project enforces: its own
// setting when it has one, the system default otherwise. A stored 0 is an
// explicit "no cooldown" and must win over the system default, which is why
// the column is nullable rather than just zero-valued.
func (u *usecase) cooldownFor(contestProjectID int32) (int32, error) {
	seconds, err := u.query.GetSubmissionCooldownSeconds(contestProjectID)
	if err != nil {
		return 0, err
	}

	if seconds == nil {
		return int32(u.config.Judge.SubmissionCooldownSeconds), nil
	}

	return *seconds, nil
}

// checkCooldown spaces out one contestant's submissions. The reservation is
// taken here rather than derived from past submissions because the submission
// record is written asynchronously — see service.SubmissionCooldown.
func (u *usecase) checkCooldown(input *Input) error {
	cooldown, err := u.cooldownFor(input.ContestProjectID)
	if err != nil {
		return err
	}

	ok, retryAfter, err := u.cooldown.Reserve(input.UserID, input.ContestProjectID, cooldown)
	if err != nil {
		return err
	}

	if !ok {
		return &SubmissionCooldownError{RetryAfterSeconds: retryAfter}
	}

	return nil
}

func (u *usecase) validate(input *Input) error {
	if input.UserID <= 0 {
		return ErrorInvalidUserID
	}
	if input.ContestProjectID <= 0 {
		return ErrorInvalidContestProjectID
	}
	if input.Project == nil {
		return ErrorProjectRequired
	}
	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to submit user project", "userId", input.UserID, "projectId", input.ContestProjectID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if err := u.query.CheckUserIsParticipant(input.ContestProjectID, input.UserID); err != nil {
		u.logger.Errorf("User is not a participant", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if err := u.query.CheckProjectAndGradingItemsExists(input.ContestProjectID, input.GradingItemIDs); err != nil {
		u.logger.Errorf("Failed to check project and grading items", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	// After the existence checks (so a missing contest project keeps its
	// original 404 message), before the cooldown (see below).
	if err := u.checkSubmissionWindow(input); err != nil {
		u.logger.Errorf("Submission rejected by the eligibility window", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	// Last of the gates, so a submission that would be rejected anyway is not
	// held against the contestant's cooldown.
	if err := u.checkCooldown(input); err != nil {
		return nil, err
	}

	if err := u.publisher.Publish(
		input.UserID,
		input.ContestProjectID,
		len(input.GradingItemIDs) != 0,
		input.GradingItemIDs,
		input.Project,
		input.DynamicFields,
	); err != nil {
		// Validation now happens before the response, so an invalid submission
		// is a 400 the contestant sees, not a 200 followed by a submission that
		// never appears.
		u.logger.Errorf("Failed to publish submission", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, ErrorInvalidData
	}

	u.logger.Infof("Successfully submitted user project", "userId", input.UserID, "projectId", input.ContestProjectID)
	return &Output{}, nil
}
