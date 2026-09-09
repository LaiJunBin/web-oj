package lifecycle

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle/command"
)

var (
	ErrorInvalidCommand = errors.New("command is invalid, must not be empty")
)

type Command struct {
	name        command.Name
	description string
	text        string
}

type CommandConfig struct {
	Name        string
	Description string
	Text        string
}

func NewCommand(config *CommandConfig) (Command, error) {
	name, err := command.NewName(config.Name)
	if err != nil {
		return Command{}, err
	}

	return Command{
		name:        name,
		description: config.Description,
		text:        config.Text,
	}, nil
}

func MustNewCommand(config *CommandConfig) Command {
	cmd, err := NewCommand(config)
	if err != nil {
		panic(err)
	}

	return cmd
}
