package setstaffnote

import (
	"errors"
	"net/http"

	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"judge-backend/internal/app/shared/types/httperror"
)

var (
	ErrorJudgmentIDIsRequired = httperror.New(http.StatusBadRequest, "Judgment ID is required")
	ErrorJudgmentNotFound     = httperror.New(http.StatusNotFound, "Judgment not found")
)

type Input struct {
	JudgmentID string
	// StaffNote is the note as typed by the operator. The empty string means
	// "clear the note": the API has no separate delete, and a retracted note
	// must disappear from the contestant's view.
	StaffNote string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.JudgmentRepository
	service    service.JudgmentService
}

func NewUsecase(
	repository repository.JudgmentRepository,
	service service.JudgmentService,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.JudgmentID == "" {
		return ErrorJudgmentIDIsRequired
	}
	return nil
}

// Execute accepts the note for any judgment status: a note is commentary
// on the judgment, not part of its lifecycle, so it may be written while
// the judgment is still queued or judging and long after it finished.
// Last write wins; there is no author or history.
func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	judgment, err := u.repository.FindByID(input.JudgmentID)
	if err != nil {
		if errors.Is(err, repository.ErrJudgmentNotFound) {
			return nil, ErrorJudgmentNotFound
		}
		return nil, err
	}

	var note *string
	if input.StaffNote != "" {
		value := input.StaffNote
		note = &value
	}

	u.service.RequestUpdateStaffNoteByID(judgment.ID(), note)

	return &Output{}, nil
}
