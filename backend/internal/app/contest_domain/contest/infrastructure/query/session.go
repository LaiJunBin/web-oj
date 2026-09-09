package query

import (
	"context"
	"fmt"
	"judge-backend/internal/app/contest_domain/contest/application/query"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/redis/go-redis/v9"
)

type SessionQuery struct {
	rdb    *redis.Client
	logger sharedtypes.Logger
}

func NewSessionQuery(
	rdb *redis.Client,
	logger sharedtypes.Logger,
) query.SessionQuery {
	return &SessionQuery{
		rdb:    rdb,
		logger: logger,
	}
}

func (q *SessionQuery) GetTTL(sessionID string) int {
	q.logger.Debugf("Starting to get session TTL", "sessionId", sessionID)

	ctx := context.Background()
	key := fmt.Sprintf("session_%s", sessionID)
	ttl := int(q.rdb.TTL(ctx, key).Val().Seconds())

	q.logger.Debugf("Successfully got session TTL", "sessionId", sessionID, "ttl", ttl)
	return ttl
}
