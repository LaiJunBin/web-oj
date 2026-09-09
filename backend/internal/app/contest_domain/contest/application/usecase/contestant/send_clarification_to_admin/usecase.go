package sendclarificationtoadmin

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorSubjectIsRequired = httperror.New(http.StatusBadRequest, "subject is required")
	ErrorContentIsRequired = httperror.New(http.StatusBadRequest, "content is required")
)

type Input struct {
	SenderUserID     int32
	ContestProjectID int32
	Subject          string
	Content          string
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository          repository.ClarificationRepository
	contestProjectQuery query.ContestProjectQuery
	logger              sharedtypes.Logger
}

func NewUsecase(
	repository repository.ClarificationRepository,
	contestProjectQuery query.ContestProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		repository:          repository,
		contestProjectQuery: contestProjectQuery,
		logger:              logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Subject == "" {
		return ErrorSubjectIsRequired
	}

	if input.Content == "" {
		return ErrorContentIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to send clarification to admin", "userId", input.SenderUserID, "projectId", input.ContestProjectID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.SenderUserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if err := u.contestProjectQuery.CheckUserCanAccessProject(input.ContestProjectID, input.SenderUserID); err != nil {
		u.logger.Errorf("User cannot access project", "userId", input.SenderUserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	clarification, err := clarification.New(&clarification.ClarificationConfig{
		SenderUserID:     input.SenderUserID,
		ContestProjectID: input.ContestProjectID,
		Subject:          input.Subject,
		Content:          input.Content,
	})
	if err != nil {
		u.logger.Errorf("Failed to create clarification", "userId", input.SenderUserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if err := u.repository.Save(clarification); err != nil {
		u.logger.Errorf("Failed to save clarification", "userId", input.SenderUserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully sent clarification to admin", "userId", input.SenderUserID, "projectId", input.ContestProjectID, "clarificationId", clarification.ID())
	return &Output{
		ID: clarification.ID(),
	}, nil
}
