package event

import "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"

type CancelBuildEvent struct {
	build *build.Build
}

func (c *CancelBuildEvent) Topic() string {
	return "cancel_build"
}

func NewCancelBuildEvent(build *build.Build) *CancelBuildEvent {
	return &CancelBuildEvent{
		build: build,
	}
}

func (c *CancelBuildEvent) Build() *build.Build {
	return c.build
}
