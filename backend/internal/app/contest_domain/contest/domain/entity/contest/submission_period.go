package contest

import (
	"judge-backend/internal/app/contest_domain/contest/domain/common"
	"time"
)

type SubmissionPeriod struct {
	common.DurationPeriod
}

func NewSubmissionPeriod(startTime, endTime time.Time) (SubmissionPeriod, error) {
	dp, err := common.NewDurationPeriod(startTime, endTime)
	if err != nil {
		return SubmissionPeriod{}, err
	}

	return SubmissionPeriod{DurationPeriod: dp}, nil
}

func (sp SubmissionPeriod) IsSubmissionPeriod(currentTime time.Time) bool {
	return sp.IsWithin(currentTime)
}
