package group

import (
	"errors"
)

type Group struct {
	id          int
	name        string
	description string
}

type GroupConfig struct {
	Name        string
	Description string
}

func New(config *GroupConfig) (*Group, error) {
	if config.Name == "" {
		return nil, errors.New("group name cannot be empty")
	}

	return &Group{
		id:          0,
		name:        config.Name,
		description: config.Description,
	}, nil
}

func (g *Group) GetID() int {
	return g.id
}

func (g *Group) SetID(id int) {
	g.id = id
}

func (g *Group) GetName() string {
	return g.name
}

func (g *Group) SetName(name string) error {
	if name == "" {
		return errors.New("group name cannot be empty")
	}
	g.name = name
	return nil
}

func (g *Group) GetDescription() string {
	return g.description
}

func (g *Group) SetDescription(description string) {
	g.description = description
}
