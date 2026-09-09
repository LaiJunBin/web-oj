package rejudgecontestproject

import (
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidContestProjectID        = httperror.New(http.StatusBadRequest, "Invalid contest project ID")
	ErrorRejudgeAlreadyAwaitingDecision = httperror.New(http.StatusBadRequest, "One or more submissions already have a rejudge awaiting decision")
	ErrorRequiredReason                 = httperror.New(http.StatusBadRequest, "Required reason")
)

type Input struct {
	ContestProjectID int32
	GradingItemIds   *[]int32
	Reason           string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository        repository.SubmissionRepository
	rejudgeRepository repository.RejudgeRepository
	service           service.SubmissionService
}

func NewUsecase(
	repository repository.SubmissionRepository,
	rejudgeRepository repository.RejudgeRepository,
	service service.SubmissionService,
) Usecase {
	return &usecase{
		repository:        repository,
		rejudgeRepository: rejudgeRepository,
		service:           service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ContestProjectID <= 0 {
		return ErrorInvalidContestProjectID
	}

	if input.Reason == "" {
		return ErrorRequiredReason
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	gradingItems := make([]int32, 0)
	if input.GradingItemIds != nil {
		gradingItems = *input.GradingItemIds
	}

	submissions, err := u.repository.FindByContestProjectIDAndGradingItemIDs(input.ContestProjectID, gradingItems)
	if err != nil {
		return nil, err
	}
	// Refuse to stack a second rejudge on submissions whose previous one is still
	// undecided. Two outstanding rejudges can be approved in either order, and
	// approving the older one afterwards silently replaces the newer result with a
	// staler score. Cancelling the outstanding rejudge is the way to change course.
	submissionIDs := make([]string, 0, len(submissions))
	for _, submission := range submissions {
		submissionIDs = append(submissionIDs, submission.ID())
	}
	awaiting, err := u.rejudgeRepository.FindAwaitingDecisionBySubmissionIDs(submissionIDs)
	if err != nil {
		return nil, err
	}
	if len(awaiting) > 0 {
		return nil, ErrorRejudgeAlreadyAwaitingDecision
	}

	u.service.RequestRejudgeSubmissions(input.ContestProjectID, submissions, input.Reason)

	return &Output{}, nil
}
