package contest

import (
	"judge-backend/internal/app/contest_domain/contest/domain/common"
	"time"
)

type ActivationPeriod struct {
	common.DurationPeriod
}

func NewActivationPeriod(startTime, endTime time.Time) (ActivationPeriod, error) {
	dp, err := common.NewDurationPeriod(startTime, endTime)
	if err != nil {
		return ActivationPeriod{}, err
	}

	return ActivationPeriod{DurationPeriod: dp}, nil
}

func (ap ActivationPeriod) IsActive(currentTime time.Time) bool {
	return ap.IsWithin(currentTime)
}
