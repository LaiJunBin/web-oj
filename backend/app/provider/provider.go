package provider

import (
	"judge-backend/api"
	"judge-backend/app"
	"judge-backend/database"
	activity "judge-backend/internal/app/contest_domain/activity"
	"judge-backend/internal/app/contest_domain/contest"
	contestinfraprovider "judge-backend/internal/app/contest_domain/contest/infrastructure/provider"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment"
	"judge-backend/internal/app/identity_domain/auth"
	"judge-backend/internal/app/identity_domain/auth/adapter/middleware"
	authrepository "judge-backend/internal/app/identity_domain/auth/domain/repository"
	"judge-backend/internal/app/identity_domain/user"
	"judge-backend/internal/app/identity_domain/user/application/service/user/roleconverter"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	userservice "judge-backend/internal/app/identity_domain/user/domain/service"
	"judge-backend/internal/app/judge_domain/judge"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	judgeworker "judge-backend/internal/app/judge_domain/worker"
	portabilitybundle "judge-backend/internal/app/portability_domain/bundle"
	problem "judge-backend/internal/app/project_domain/project"
	projectservice "judge-backend/internal/app/project_domain/project/infrastructure/service"
	"judge-backend/internal/app/shared"
	sharedmiddleware "judge-backend/internal/app/shared/adapter/middleware"
	"judge-backend/internal/app/shared/rolebits"
	"judge-backend/internal/app/shared/types/provider"
	"judge-backend/internal/config"
	"reflect"
	"runtime"
	"sort"

	"go.uber.org/fx"
)

var (
	Middlewares = []interface{}{
		middleware.Csrf,
		middleware.Authorization,
		middleware.WithRefreshToken,
		middleware.WithUserID,
		middleware.AuthAdmin,
		middleware.AuthContestant,
		middleware.AuthRead,
	}

	Databases = []interface{}{
		database.NewRedisStore,
		database.DatabaseProvider,
		database.NewRedisDB,
		database.NewMongoDB,
	}

	Providers = []*provider.Provider{
		shared.NewProvider(),
		auth.NewProvider(),
		user.NewProvider(),
		execution_environment.NewProvider(),
		problem.NewProvider(),
		activity.NewProvider(),
		judgeworker.NewProvider(),
		contest.NewProvider(),
		judge.NewProvider(),
		portabilitybundle.NewProvider(),
	}

	Invokes = []interface{}{}
)

func GetMiddlewares(middlewares []api.MiddlewareFunc) []api.MiddlewareFunc {
	orders := make(map[string]int)
	for i, m := range Middlewares {
		orders[runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()+".func1"] = i
	}

	sort.Slice(middlewares, func(i, j int) bool {
		a := runtime.FuncForPC(reflect.ValueOf(middlewares[i]).Pointer()).Name()
		b := runtime.FuncForPC(reflect.ValueOf(middlewares[j]).Pointer()).Name()
		return orders[a] < orders[b]
	})

	return middlewares
}

func GetProvides() []interface{} {
	provides := []interface{}{
		// Bridge: contest BC uses JudgeAvailableStrategiesProvider interface; bind judge's provider here.
		func(p prioritystrategies.Provider) contestinfraprovider.JudgeAvailableStrategiesProvider { return p },
		// Bridge: resolve valid parser names at composition root so project_domain stays decoupled from judge_domain.
		func() projectservice.ValidParserNames {
			return projectservice.ValidParserNames(resultparser.GetValidParserNames())
		},
		// Bridge: compute all role bitmasks once at composition root into a single injectable struct.
		provideRoleMasks,
		// Bridge: user BC needs sessions revoked when a password changes; the
		// token store belongs to auth, so bind its repository here rather than
		// let one identity BC import the other's.
		func(r authrepository.RefreshTokenRepository) userservice.SessionRevoker {
			return sessionRevoker{tokens: r}
		},
		// The per-(IP,username) login limiter, provided as ONE shared instance so
		// the middleware that enforces it and the password-reset usecase that
		// clears a user's record act on the same buckets. Sized from config;
		// non-positive knobs make it a pass-through (see NewLoginRateLimiter).
		newSharedLoginLimiter,
		// Bridge: user BC clears a contestant's login throttling on password
		// reset; the limiter lives in the shared transport layer, so bind it here.
		func(l *sharedmiddleware.LoginRateLimiter) userservice.LoginAttemptForgetter {
			return l
		},
		fx.Annotate(GetMiddlewares, fx.ParamTags(`group:"middleware"`)),
		fx.Annotate(app.NewServer, fx.ParamTags(`group:"controller"`, `group:"handler"`)),
		// http.NewHttpClient,
	}

	// Bridge: the composite bundle spans three bounded contexts, so all of
	// its ports are bound here — see bundle_bridges.go.
	provides = append(provides, bundleBridges()...)

	middlewares := []interface{}{}
	for _, m := range Middlewares {
		middlewares = append(middlewares, fx.Annotate(m, fx.ResultTags(`group:"middleware"`)))
	}

	provides = append(provides, Databases...)
	provides = append(provides, middlewares...)

	for _, p := range Providers {
		controllers := []interface{}{}
		for _, c := range p.Config.Controllers {
			controllers = append(controllers, fx.Annotate(c, fx.As(new(app.Controller)), fx.ResultTags(`group:"controller"`)))
		}

		provides = append(provides, p.Config.Infrastructures...)
		provides = append(provides, p.Config.Presenters...)
		provides = append(provides, p.Config.Services...)
		provides = append(provides, p.Config.Usecases...)
		provides = append(provides, p.Config.Applications...)
		provides = append(provides, p.Config.Factories...)
		provides = append(provides, p.Config.Providers...)
		provides = append(provides, controllers...)

		Invokes = append(Invokes, p.Config.Handlers...)
		Invokes = append(Invokes, p.Config.Invokes...)
	}

	return provides
}

func GetInvokes() []interface{} {
	return Invokes
}

// newSharedLoginLimiter builds the per-(IP,username) login limiter from config.
// A non-positive burst or per-minute makes it a pass-through (see
// NewLoginRateLimiter); when it is enabled, SetupHandler installs the middleware
// that uses this same instance.
func newSharedLoginLimiter(cfg *config.Config) *sharedmiddleware.LoginRateLimiter {
	burst, perMinute := 0, 0
	if cfg.Server.LoginRateLimitBurst != nil {
		burst = *cfg.Server.LoginRateLimitBurst
	}
	if cfg.Server.LoginRateLimitPerMinute != nil {
		perMinute = *cfg.Server.LoginRateLimitPerMinute
	}
	return sharedmiddleware.NewLoginRateLimiter(burst, perMinute)
}

func provideRoleMasks(rc roleconverter.RoleConverter) rolebits.RoleMasks {
	return rolebits.RoleMasks{
		Admin:      rolebits.AdminRoleMask(rc.ToBit(role.Roles{role.RoleAdmin})),
		Contestant: rolebits.ContestantRoleMask(rc.ToBit(role.Roles{role.RoleContestant})),
		APIReader:  rolebits.APIReaderRoleMask(rc.ToBit(role.Roles{role.RoleAPIReader})),
	}
}

// sessionRevoker adapts the auth token store to the user BC's SessionRevoker.
// It exists so the two contexts share an interface rather than a package.
type sessionRevoker struct {
	tokens authrepository.RefreshTokenRepository
}

func (s sessionRevoker) RevokeAllSessions(userID int) error {
	return s.tokens.DeleteAllByUserID(userID)
}
