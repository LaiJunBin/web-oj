package asset

import (
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"time"
)

type Asset struct {
	id           ID
	projectID    project.ID
	displayName  Name
	size         Size
	lastModified LastModified
	enabled      bool

	sharedtypes.TimeStamps
}

type AssetConfig struct {
	ID          string
	ProjectID   project.ID
	DisplayName string
	Size        int64
	LastMod     time.Time
	Enabled     bool
}

func New(config *AssetConfig) (*Asset, error) {
	id, err := NewID(config.ID)
	if err != nil {
		return nil, err
	}

	displayName, err := NewName(config.DisplayName)
	if err != nil {
		return nil, err
	}

	size := NewSize(config.Size)

	lastModified, err := NewLastModified(config.LastMod)
	if err != nil {
		return nil, err
	}

	return &Asset{
		id:           id,
		projectID:    config.ProjectID,
		displayName:  displayName,
		size:         size,
		lastModified: lastModified,
		enabled:      config.Enabled,
	}, nil
}

func (a *Asset) ProjectID() project.ID {
	return a.projectID
}

func (a *Asset) ID() ID {
	return a.id
}

func (a *Asset) DisplayName() Name {
	return a.displayName
}

func (a *Asset) Size() Size {
	return a.size
}

func (a *Asset) LastModified() LastModified {
	return a.lastModified
}

func (a *Asset) IsEnabled() bool {
	return a.enabled
}

func (a *Asset) Enable() {
	a.enabled = true
}

func (a *Asset) Disable() {
	a.enabled = false
}

func (a *Asset) SetIsEnabled(enabled bool) {
	a.enabled = enabled
}

func (a *Asset) SetID(id string) error {
	i, err := NewID(id)
	if err != nil {
		return err
	}
	a.id = i
	return nil
}

func (a *Asset) SetDisplayName(displayName string) error {
	name, err := NewName(displayName)
	if err != nil {
		return err
	}

	a.displayName = name
	return nil
}

func (a *Asset) SetSize(size int64) {
	a.size = NewSize(size)
}

func (a *Asset) SetLastModified(lastModified time.Time) error {
	modified, err := NewLastModified(lastModified)
	if err != nil {
		return err
	}

	a.lastModified = modified
	return nil
}
