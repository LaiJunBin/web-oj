package setup

import (
	"fmt"
	"judge-backend/app"
	"judge-backend/app/provider"
	"judge-backend/internal/app/identity_domain/auth/adapter/middleware"
	dockerruntime "judge-backend/internal/app/judge_domain/judge/infrastructure/runtime/docker"
	fakeruntime "judge-backend/internal/app/judge_domain/judge/infrastructure/runtime/fake"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"
	"reflect"
	"runtime"

	"go.uber.org/fx"
)

func (s *setup) setupServer() {
	fmt.Println("Setting up server...")
	ignoreMiddlewareProvides := []interface{}{
		middleware.Csrf,
	}

	for _, ignore := range ignoreMiddlewareProvides {
		b := runtime.FuncForPC(reflect.ValueOf(ignore).Pointer()).Name()
		for i, middleware := range provider.Middlewares {
			a := runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name()
			if a == b {
				provider.Middlewares = append(provider.Middlewares[:i], provider.Middlewares[i+1:]...)
				break
			}
		}
	}

	defaultProvides := provider.GetProvides()
	ignoreProvides := []interface{}{
		logger.NewZapLogger,
		// Judgments run against the in-memory fake runtime in feature tests so
		// the full submission → judged pipeline is deterministic and needs no
		// Docker daemon or pre-pulled images.
		dockerruntime.NewRuntime,
	}

	for _, ignore := range ignoreProvides {
		ignorePointer := reflect.ValueOf(ignore)
		for i, provide := range defaultProvides {
			providePointer := reflect.ValueOf(provide)
			if providePointer == ignorePointer {
				defaultProvides = append(defaultProvides[:i], defaultProvides[i+1:]...)
				break
			}
		}
	}

	provides := append([]interface{}{
		config.NewFeatureConfig,
		logger.NewMockLogger,
		fakeruntime.NewRuntime,
	}, defaultProvides...)
	// RegisterJudgeTestSeam must come before app.SetupHandler: SetupHandler
	// blocks in gin's Run, so invokes after it never execute.
	invokes := append(provider.GetInvokes(), RegisterJudgeTestSeam, app.SetupHandler)

	fx.New(
		fx.Provide(
			provides...,
		),
		fx.Invoke(
			invokes...,
		),
	).Run()
}
