package buildquery

import "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query/models/buildinfo"

type BuildQueryService interface {
	GetBuilds(tag string) ([]buildinfo.BuildInfo, error)
}
