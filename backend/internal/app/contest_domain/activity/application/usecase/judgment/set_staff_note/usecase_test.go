package setstaffnote_test

import (
	"errors"
	"testing"

	setstaffnote "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/set_staff_note"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/judgment"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func strPtr(s string) *string { return &s }

func TestUsecase_Execute(t *testing.T) {
	t.Run("sets the note on a finished judgment", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		repo.On("FindByID", "judgment-1").Return(judgment.New("judgment-1", "submission-1", "Judged"), nil).Once()
		svc.On("RequestUpdateStaffNoteByID", "judgment-1", strPtr("第 2 題輸出多了空白")).Return().Once()

		output, err := uc.Execute(&setstaffnote.Input{JudgmentID: "judgment-1", StaffNote: "第 2 題輸出多了空白"})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	// A note is commentary, not lifecycle: it may be written while the
	// judgment is still running (unlike cancel, which gates on status).
	t.Run("accepts a note while the judgment is still judging", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		repo.On("FindByID", "judgment-2").Return(judgment.New("judgment-2", "submission-2", "Judging"), nil).Once()
		svc.On("RequestUpdateStaffNoteByID", "judgment-2", strPtr("note")).Return().Once()

		_, err := uc.Execute(&setstaffnote.Input{JudgmentID: "judgment-2", StaffNote: "note"})

		assert.NoError(t, err)
	})

	// The empty string is the only way to retract a note through the API,
	// and it must reach the judge domain as nil, not as "".
	t.Run("an empty string clears the note", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		repo.On("FindByID", "judgment-1").Return(judgment.New("judgment-1", "submission-1", "Judged"), nil).Once()
		svc.On("RequestUpdateStaffNoteByID", "judgment-1", (*string)(nil)).Return().Once()

		_, err := uc.Execute(&setstaffnote.Input{JudgmentID: "judgment-1", StaffNote: ""})

		assert.NoError(t, err)
	})

	// Last write wins: a second note is passed through verbatim, with no
	// merge and no knowledge of what was there before.
	t.Run("a later note overwrites without merging", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		repo.On("FindByID", "judgment-1").Return(judgment.New("judgment-1", "submission-1", "Judged"), nil).Twice()
		svc.On("RequestUpdateStaffNoteByID", "judgment-1", strPtr("first")).Return().Once()
		svc.On("RequestUpdateStaffNoteByID", "judgment-1", strPtr("second")).Return().Once()

		_, err := uc.Execute(&setstaffnote.Input{JudgmentID: "judgment-1", StaffNote: "first"})
		assert.NoError(t, err)
		_, err = uc.Execute(&setstaffnote.Input{JudgmentID: "judgment-1", StaffNote: "second"})
		assert.NoError(t, err)
	})

	t.Run("rejects an empty judgment id before touching the repository", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		_, err := uc.Execute(&setstaffnote.Input{JudgmentID: "", StaffNote: "note"})

		assert.ErrorIs(t, err, setstaffnote.ErrorJudgmentIDIsRequired)
	})

	t.Run("an unknown judgment is a 404, and nothing is published", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		repo.On("FindByID", "missing").Return(nil, repository.ErrJudgmentNotFound).Once()

		_, err := uc.Execute(&setstaffnote.Input{JudgmentID: "missing", StaffNote: "note"})

		assert.ErrorIs(t, err, setstaffnote.ErrorJudgmentNotFound)
		svc.AssertNotCalled(t, "RequestUpdateStaffNoteByID", mock.Anything, mock.Anything)
	})

	t.Run("any other repository failure is passed through", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := setstaffnote.NewUsecase(repo, svc)

		boom := errors.New("mongo unavailable")
		repo.On("FindByID", "judgment-1").Return(nil, boom).Once()

		_, err := uc.Execute(&setstaffnote.Input{JudgmentID: "judgment-1", StaffNote: "note"})

		assert.ErrorIs(t, err, boom)
	})
}
