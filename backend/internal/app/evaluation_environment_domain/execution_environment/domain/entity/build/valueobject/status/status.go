package buildstatus

import "errors"

var (
	ErrInvalidStatus = errors.New("invalid status")
)

var (
	PENDING   = Status{name: "PENDING", value: 0, successor: []int{1, 2, 4}}
	BUILDING  = Status{name: "BUILDING", value: 1, successor: []int{2, 3, 4}}
	CANCELLED = Status{name: "CANCELLED", value: 2, successor: []int{0}}
	SUCCESS   = Status{name: "SUCCESS", value: 3, successor: []int{0}}
	ERROR     = Status{name: "ERROR", value: 4, successor: []int{0}}
)

type Status struct {
	name      string
	value     int
	successor []int
}

func New(value int) (Status, error) {
	switch value {
	case 0:
		return PENDING, nil
	case 1:
		return BUILDING, nil
	case 2:
		return CANCELLED, nil
	case 3:
		return SUCCESS, nil
	case 4:
		return ERROR, nil
	default:
		return Status{}, ErrInvalidStatus
	}
}

func NewByName(name string) (Status, error) {
	switch name {
	case "PENDING":
		return PENDING, nil
	case "BUILDING":
		return BUILDING, nil
	case "CANCELLED":
		return CANCELLED, nil
	case "SUCCESS":
		return SUCCESS, nil
	case "ERROR":
		return ERROR, nil
	default:
		return Status{}, ErrInvalidStatus
	}
}

func (s Status) Value() int {
	return s.value
}

func (s Status) Equals(other Status) bool {
	return s.value == other.value && s.name == other.name
}

func (s Status) String() string {
	return s.name
}

func (s Status) CanChangeTo(other Status) bool {
	for _, v := range s.successor {
		if v == other.value {
			return true
		}
	}
	return false
}
