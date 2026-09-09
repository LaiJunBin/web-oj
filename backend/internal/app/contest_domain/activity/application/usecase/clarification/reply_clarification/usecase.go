package replyclarification

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"net/http"
)

var (
	ErrorInvalidClarificationID = httperror.New(http.StatusBadRequest, "invalid clarification ID")
	ErrorContentIsRequired      = httperror.New(http.StatusBadRequest, "content is required")
	ErrorClarificationNotFound  = httperror.New(http.StatusNotFound, "clarification not found")
)

type Input struct {
	ID           int32
	SenderUserID int32
	Subject      *string
	Content      string
	SendToAll    bool
}

type Output struct {
	ID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ClarificationRepository
	service    service.ClarificationService
}

func NewUsecase(
	repository repository.ClarificationRepository,
	service service.ClarificationService,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidClarificationID
	}

	if input.Content == "" {
		return ErrorContentIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	c, err := u.repository.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	if c == nil {
		return nil, ErrorClarificationNotFound
	}

	replyClarification, err := c.Reply(&clarification.ReplyConfig{
		SenderUserID: input.SenderUserID,
		Subject:      input.Subject,
		Content:      input.Content,
		SendToAll:    input.SendToAll,
	})

	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(replyClarification); err != nil {
		return nil, err
	}

	u.service.NotifyClarificationToUser(replyClarification)

	return &Output{
		ID: replyClarification.ID(),
	}, nil
}
