package dto

import (
	"time"
)

type Project struct {
	ID          int32
	Name        string
	Description string
	Image       *string
	TestImage   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectDetail struct {
	ID            int32
	Name          string
	Description   string
	Image         *Image
	TestImage     *Image
	Env           string
	TestEnv       string
	ResultParser  string
	Assets        []*Asset
	DynamicFields []*DynamicField
	Lifecycle     Lifecycle
	TestLifecycle Lifecycle
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
