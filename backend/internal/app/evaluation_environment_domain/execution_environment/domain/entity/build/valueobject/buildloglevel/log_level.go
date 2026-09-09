package buildloglevel

import (
	"fmt"
)

var (
	Info    = Level{value: 0, text: "info"}
	Warning = Level{value: 1, text: "warning"}
	Success = Level{value: 2, text: "success"}
	Error   = Level{value: 3, text: "error"}
)

type Level struct {
	value int
	text  string
}

func New(value int) (Level, error) {
	switch value {
	case 0:
		return Info, nil
	case 1:
		return Warning, nil
	case 2:
		return Success, nil
	case 3:
		return Error, nil
	default:
		return Level{}, fmt.Errorf("invalid log level: %d", value)
	}
}

func NewByText(text string) (Level, error) {
	switch text {
	case "info":
		return Info, nil
	case "warning":
		return Warning, nil
	case "success":
		return Success, nil
	case "error":
		return Error, nil
	default:
		return Level{}, fmt.Errorf("invalid log level text: %s", text)
	}
}

func (l Level) String() string {
	return l.text
}

func (l Level) Value() int {
	return l.value
}
