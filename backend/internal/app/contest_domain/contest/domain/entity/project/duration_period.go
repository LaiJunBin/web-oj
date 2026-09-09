package project

import (
	"errors"
	"time"
)

var (
	ErrorStartTimeIsRequired         = errors.New("start time is required")
	ErrorEndTimeIsRequired           = errors.New("end time is required")
	ErrorEndTimeMustBeAfterStartTime = errors.New("end time must be after start time")
)

type DurationPeriod struct {
	startTime time.Time
	endTime   time.Time
}

func NewDurationPeriod(startTime, endTime time.Time) (DurationPeriod, error) {
	dp := DurationPeriod{
		startTime: startTime,
		endTime:   endTime,
	}

	if err := dp.validate(); err != nil {
		return DurationPeriod{}, err
	}

	return dp, nil
}

func (dp DurationPeriod) validate() error {
	if dp.startTime.IsZero() {
		return ErrorStartTimeIsRequired
	}

	if dp.endTime.IsZero() {
		return ErrorEndTimeIsRequired
	}

	if dp.endTime.Before(dp.startTime) {
		return ErrorEndTimeMustBeAfterStartTime
	}

	return nil
}

func (dp DurationPeriod) IsWithin(currentTime time.Time) bool {
	return currentTime.After(dp.startTime) && currentTime.Before(dp.endTime)
}

func (dp DurationPeriod) StartTime() time.Time {
	return dp.startTime
}

func (dp DurationPeriod) EndTime() time.Time {
	return dp.endTime
}

func (dp DurationPeriod) TimeLeftToStart(currentTime time.Time) time.Duration {
	return dp.startTime.Sub(currentTime)
}

func (dp DurationPeriod) TimeLeftToEnd(currentTime time.Time) time.Duration {
	return dp.endTime.Sub(currentTime)
}
