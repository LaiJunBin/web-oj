package command

import "errors"

var (
	ErrorInvalidCommand = errors.New("command is invalid, must not be empty")
)

type Command struct {
	value string
}

func New(value string) (Command, error) {
	command := Command{value: value}
	if err := command.validate(); err != nil {
		return Command{}, err
	}

	return command, nil
}

func MustNew(value string) Command {
	command, err := New(value)
	if err != nil {
		panic(err)
	}

	return command
}

func (n Command) validate() error {
	if n.value == "" {
		return ErrorInvalidCommand
	}

	return nil
}

func (n Command) String() string {
	return n.value
}

func (n Command) Equals(other Command) bool {
	return n.value == other.value
}
