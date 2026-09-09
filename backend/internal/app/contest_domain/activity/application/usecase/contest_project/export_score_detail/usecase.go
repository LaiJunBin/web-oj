package exportscoredetail

import (
	"errors"
	"fmt"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"judge-backend/internal/app/shared/types/httperror"
	"time"
)

var (
	ErrorContestProjectNotFound = errors.New("contest project not found")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	ContestProjectID int32
}

type Output struct {
	Filename string
	Content  []byte
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query     query.ContestProjectQuery
	generator service.ScoreDetailReportGenerator
}

func NewUsecase(
	query query.ContestProjectQuery,
	generator service.ScoreDetailReportGenerator,
) Usecase {
	return &usecase{
		query:     query,
		generator: generator,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	cp, err := u.query.GetContestProjectByID(input.ContestProjectID)
	if err != nil {
		if e, ok := err.(*httperror.Error); ok && e.Code == 404 {
			return nil, ErrorContestProjectNotFound
		}
		return nil, ErrorInternalServer
	}

	contestants, err := u.query.GetContestantScoreDetails(input.ContestProjectID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	reports := make([]*service.ContestantScoreReport, 0, len(contestants))
	for _, c := range contestants {
		items := make([]*service.ScoreReportItem, 0, len(c.Items))
		for _, item := range c.Items {
			items = append(items, &service.ScoreReportItem{
				Label:       item.Label,
				Name:        item.Name,
				Description: item.Description,
				Score:       item.Score,
				TotalScore:  item.TotalScore,
				Status:      item.Status,
			})
		}
		reports = append(reports, &service.ContestantScoreReport{
			UserID:   c.User.ID,
			Name:     c.User.Name,
			Username: c.User.Username,
			Items:    items,
		})
	}

	content, err := u.generator.Generate(reports)
	if err != nil {
		return nil, ErrorInternalServer
	}

	filename := fmt.Sprintf("%s-%s-score-detail.xlsx",
		cp.Name,
		time.Now().Format("20060102-150405"),
	)

	return &Output{Filename: filename, Content: content}, nil
}
