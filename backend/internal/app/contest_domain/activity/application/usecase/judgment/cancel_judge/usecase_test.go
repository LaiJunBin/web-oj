package canceljudge_test

import (
	canceljudge "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/cancel_judge"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/judgment"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	t.Run("should cancel queued judgment successfully", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := canceljudge.NewUsecase(repo, svc)

		j := judgment.New("judgment-1", "submission-1", "Queued")
		repo.On("FindByID", "judgment-1").Return(j, nil).Once()
		svc.On("RequestCancelJudgeByID", "judgment-1", "評測已由工作人員取消").Return().Once()

		output, err := uc.Execute(&canceljudge.Input{
			JudgmentID: "judgment-1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should cancel judging judgment successfully", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := canceljudge.NewUsecase(repo, svc)

		j := judgment.New("judgment-2", "submission-2", "Judging")
		repo.On("FindByID", "judgment-2").Return(j, nil).Once()
		svc.On("RequestCancelJudgeByID", "judgment-2", "評測已由工作人員取消").Return().Once()

		output, err := uc.Execute(&canceljudge.Input{
			JudgmentID: "judgment-2",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should use custom reason when provided", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := canceljudge.NewUsecase(repo, svc)

		j := judgment.New("judgment-1", "submission-1", "Queued")
		repo.On("FindByID", "judgment-1").Return(j, nil).Once()
		svc.On("RequestCancelJudgeByID", "judgment-1", "Custom reason").Return().Once()

		output, err := uc.Execute(&canceljudge.Input{
			JudgmentID: "judgment-1",
			Reason:     "Custom reason",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when judgment ID is empty", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := canceljudge.NewUsecase(repo, svc)

		_, err := uc.Execute(&canceljudge.Input{
			JudgmentID: "",
		})

		assert.ErrorIs(t, err, canceljudge.ErrorJudgmentIDIsRequired)
	})

	t.Run("should return error when FindByID fails", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := canceljudge.NewUsecase(repo, svc)

		repo.On("FindByID", "judgment-1").Return((*judgment.Judgment)(nil), assert.AnError).Once()

		_, err := uc.Execute(&canceljudge.Input{JudgmentID: "judgment-1"})

		assert.Error(t, err)
	})

	t.Run("should return error when judgment cannot be cancelled", func(t *testing.T) {
		repo := mocks.NewJudgmentRepository(t)
		svc := mocks.NewJudgmentService(t)
		uc := canceljudge.NewUsecase(repo, svc)

		// Judged status cannot be cancelled
		j := judgment.New("judgment-1", "submission-1", "Judged")
		repo.On("FindByID", "judgment-1").Return(j, nil).Once()

		_, err := uc.Execute(&canceljudge.Input{JudgmentID: "judgment-1"})

		assert.ErrorIs(t, err, canceljudge.ErrorJudgmentCannotBeCancelled)
	})
}
