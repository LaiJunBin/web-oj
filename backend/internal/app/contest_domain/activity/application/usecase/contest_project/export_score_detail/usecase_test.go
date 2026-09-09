package exportscoredetail_test

import (
	"strings"
	"testing"

	exportscoredetail "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/export_score_detail"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	domainservice "judge-backend/internal/app/contest_domain/activity/domain/service"
	"judge-backend/internal/app/contest_domain/activity/mocks"
	"judge-backend/internal/app/shared/types/httperror"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	contestProject := &dto.ContestProject{
		ID:   1,
		Name: "Web Design",
	}

	contestants := []*dto.ContestantScoreDetail{
		{
			User: &dto.User{ID: 1, Name: "Alice", Username: "alice"},
			Items: []*dto.ScoreDetailItem{
				{
					GradingItemID: 1,
					Label:         "1",
					Name:          "Item 1",
					Description:   "desc",
					Score:         80,
					TotalScore:    100,
					Status:        "Passed",
				},
			},
		},
	}

	expectedReports := []*domainservice.ContestantScoreReport{
		{
			UserID:   1,
			Name:     "Alice",
			Username: "alice",
			Items: []*domainservice.ScoreReportItem{
				{
					Label:       "1",
					Name:        "Item 1",
					Description: "desc",
					Score:       80,
					TotalScore:  100,
					Status:      "Passed",
				},
			},
		},
	}

	excelBytes := []byte("fake-excel-content")

	t.Run("should return excel content with generated filename", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		g := mocks.NewScoreDetailReportGenerator(t)
		uc := exportscoredetail.NewUsecase(q, g)

		q.On("GetContestProjectByID", int32(1)).Return(contestProject, nil).Once()
		q.On("GetContestantScoreDetails", int32(1)).Return(contestants, nil).Once()
		g.On("Generate", expectedReports).Return(excelBytes, nil).Once()

		output, err := uc.Execute(&exportscoredetail.Input{ContestProjectID: 1})

		assert.NoError(t, err)
		assert.Equal(t, excelBytes, output.Content)
		assert.True(t, strings.HasPrefix(output.Filename, "Web Design-"))
		assert.True(t, strings.HasSuffix(output.Filename, "-score-detail.xlsx"))
	})

	t.Run("should return error when contest project not found", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		g := mocks.NewScoreDetailReportGenerator(t)
		uc := exportscoredetail.NewUsecase(q, g)

		q.On("GetContestProjectByID", int32(99)).Return((*dto.ContestProject)(nil), httperror.New(404, "contest project not found")).Once()

		_, err := uc.Execute(&exportscoredetail.Input{ContestProjectID: 99})

		assert.ErrorIs(t, err, exportscoredetail.ErrorContestProjectNotFound)
	})

	t.Run("should return internal error when score details query fails", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		g := mocks.NewScoreDetailReportGenerator(t)
		uc := exportscoredetail.NewUsecase(q, g)

		q.On("GetContestProjectByID", int32(1)).Return(contestProject, nil).Once()
		q.On("GetContestantScoreDetails", int32(1)).Return(([]*dto.ContestantScoreDetail)(nil), assert.AnError).Once()

		_, err := uc.Execute(&exportscoredetail.Input{ContestProjectID: 1})

		assert.ErrorIs(t, err, exportscoredetail.ErrorInternalServer)
	})

	t.Run("should return internal error when generator fails", func(t *testing.T) {
		q := mocks.NewContestProjectQuery(t)
		g := mocks.NewScoreDetailReportGenerator(t)
		uc := exportscoredetail.NewUsecase(q, g)

		q.On("GetContestProjectByID", int32(1)).Return(contestProject, nil).Once()
		q.On("GetContestantScoreDetails", int32(1)).Return(contestants, nil).Once()
		g.On("Generate", expectedReports).Return(([]byte)(nil), assert.AnError).Once()

		_, err := uc.Execute(&exportscoredetail.Input{ContestProjectID: 1})

		assert.ErrorIs(t, err, exportscoredetail.ErrorInternalServer)
	})
}
