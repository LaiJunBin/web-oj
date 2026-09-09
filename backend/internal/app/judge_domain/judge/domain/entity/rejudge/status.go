package rejudge

import (
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	PENDING  = &Status{"Pending"}
	RUNNING  = &Status{"Running"}
	FINISHED = &Status{"Finished"}
	APPROVED = &Status{"Approved"}
	REJECTED = &Status{"Rejected"}
	ERROR    = &Status{"Error"}
	CANCELED = &Status{"Canceled"}

	ErrInvalidStatus = httperror.New(http.StatusBadRequest, "invalid status")
)

type Status struct {
	value string
}

func NewStatus(
	value string,
) (*Status, error) {
	switch value {
	case PENDING.value:
		return PENDING, nil
	case RUNNING.value:
		return RUNNING, nil
	case FINISHED.value:
		return FINISHED, nil
	case APPROVED.value:
		return APPROVED, nil
	case REJECTED.value:
		return REJECTED, nil
	case ERROR.value:
		return ERROR, nil
	case CANCELED.value:
		return CANCELED, nil
	default:
		return nil, ErrInvalidStatus
	}
}

func (s Status) IsPending() bool {
	return s.value == PENDING.value
}

func (s Status) IsRunning() bool {
	return s.value == RUNNING.value
}

func (s Status) IsFinished() bool {
	return s.value == FINISHED.value
}

func (s Status) IsApproved() bool {
	return s.value == APPROVED.value
}

func (s Status) IsRejected() bool {
	return s.value == REJECTED.value
}

func (s Status) IsError() bool {
	return s.value == ERROR.value
}

func (s Status) IsCanceled() bool {
	return s.value == CANCELED.value
}

func (s Status) String() string {
	return s.value
}
