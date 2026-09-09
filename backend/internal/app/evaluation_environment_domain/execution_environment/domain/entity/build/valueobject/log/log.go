package buildlog

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	"time"
)

type Log struct {
	message  string
	level buildloglevel.Level
	time  time.Time
}

func New(level buildloglevel.Level, message string, time time.Time) Log {
	return Log{
		message:  message,
		level: level,
		time:  time,
	}
}

func (l Log) Message() string {
	return l.message
}

func (l Log) Level() buildloglevel.Level {
	return l.level
}

func (l Log) Time() time.Time {
	return l.time
}
