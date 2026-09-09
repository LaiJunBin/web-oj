package sharedtypes

import "time"

type TimeStamps struct {
	createdAt time.Time
	updatedAt time.Time
}

func (ts *TimeStamps) CreatedAt() time.Time {
	return ts.createdAt
}

func (ts *TimeStamps) UpdatedAt() time.Time {
	return ts.updatedAt
}

func (ts *TimeStamps) SetCreatedAt(createdAt time.Time) {
	ts.createdAt = createdAt
}

func (ts *TimeStamps) SetUpdatedAt(updatedAt time.Time) {
	ts.updatedAt = updatedAt
}
