package logreader

import buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"

type LogReader interface {
	Read() (<-chan buildlog.Log, error)
	Close() error
}
