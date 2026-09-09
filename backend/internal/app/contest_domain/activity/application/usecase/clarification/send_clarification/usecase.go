package sendclarification

import (
	"errors"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"net/http"
)

var (
	ErrorInvalidContestProjectID             = httperror.New(http.StatusBadRequest, "invalid contest project ID")
	ErrorSubjectIsRequired                   = httperror.New(http.StatusBadRequest, "subject is required")
	ErrorReceiverUserIDOrSendToAllIsRequired = httperror.New(http.StatusBadRequest, "you must specify receiver user ID or send to all")
	ErrorReceiverUserIDIsNotExists           = httperror.New(http.StatusBadRequest, "receiver user ID is not exists")
	ErrorContestProjectNotFound              = httperror.New(http.StatusNotFound, "contest project not found")
)

type Input struct {
	ContestProjectID int32
	SenderUserID     int32
	ReceiverUserID   *int32
	Subject          string
	Content          string
	SendToAll        bool
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
	if input.ContestProjectID <= 0 {
		return ErrorInvalidContestProjectID
	}

	if input.Subject == "" {
		return ErrorSubjectIsRequired
	}

	if input.ReceiverUserID == nil && !input.SendToAll {
		return ErrorReceiverUserIDOrSendToAllIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	clarification, err := clarification.New(&clarification.ClarificationConfig{
		ContestProjectID: input.ContestProjectID,
		Subject:          input.Subject,
		Content:          input.Content,
		SenderUserID:     &input.SenderUserID,
		ReceiverUserID:   input.ReceiverUserID,
		SendToAll:        input.SendToAll,
	})

	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(clarification); err != nil {
		switch {
		case errors.Is(err, repository.ErrorContestProjectNotFound):
			return nil, ErrorContestProjectNotFound
		case errors.Is(err, repository.ErrorReceiverUserNotFound):
			return nil, ErrorReceiverUserIDIsNotExists
		default:
			return nil, err
		}
	}

	u.service.NotifyClarificationToUser(clarification)

	return &Output{
		ID: clarification.ID(),
	}, nil
}
