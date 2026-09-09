package replyclarificationtoadmin

import (
	"judge-backend/internal/app/contest_domain/contest/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorContentIsRequired      = httperror.New(http.StatusBadRequest, "content is required")
	ErrorClarificationForbidden = httperror.New(http.StatusForbidden, "you are not allowed to reply this clarification")
	ErrorClarificationNotFound  = httperror.New(http.StatusNotFound, "clarification not found")
)

type Input struct {
	UserID          int32
	ClarificationID int32
	Content         string
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ClarificationRepository
	logger     sharedtypes.Logger
}

func NewUsecase(
	repository repository.ClarificationRepository,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		repository: repository,
		logger:     logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Content == "" {
		return ErrorContentIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to reply clarification to admin", "userId", input.UserID, "clarificationId", input.ClarificationID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.UserID, "clarificationId", input.ClarificationID, "error", err)
		return nil, err
	}

	c, err := u.repository.FindByID(input.ClarificationID)
	if err != nil {
		u.logger.Errorf("Failed to find clarification", "userId", input.UserID, "clarificationId", input.ClarificationID, "error", err)
		return nil, err
	}

	if c == nil {
		u.logger.Errorf("Clarification not found", "userId", input.UserID, "clarificationId", input.ClarificationID)
		return nil, ErrorClarificationNotFound
	}

	if !c.CanReply(input.UserID) {
		u.logger.Errorf("User is not allowed to reply this clarification", "userId", input.UserID, "clarificationId", input.ClarificationID)
		return nil, ErrorClarificationForbidden
	}

	replyClarification, err := c.Reply(&clarification.ReplyConfig{
		SendUserID: input.UserID,
		Content:    input.Content,
	})
	if err != nil {
		u.logger.Errorf("Failed to create reply clarification", "userId", input.UserID, "clarificationId", input.ClarificationID, "error", err)
		return nil, err
	}

	if err := u.repository.Save(replyClarification); err != nil {
		u.logger.Errorf("Failed to save reply clarification", "userId", input.UserID, "clarificationId", input.ClarificationID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully replied clarification to admin", "userId", input.UserID, "clarificationId", input.ClarificationID, "replyId", replyClarification.ID())
	return &Output{
		ID: replyClarification.ID(),
	}, nil
}
