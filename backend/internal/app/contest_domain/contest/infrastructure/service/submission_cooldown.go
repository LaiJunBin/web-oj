package service

import (
	"context"
	"fmt"
	"time"

	"judge-backend/internal/app/contest_domain/contest/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/redis/go-redis/v9"
)

type SubmissionCooldown struct {
	rdb    *redis.Client
	logger sharedtypes.Logger
}

func NewSubmissionCooldown(
	rdb *redis.Client,
	logger sharedtypes.Logger,
) service.SubmissionCooldown {
	return &SubmissionCooldown{
		rdb:    rdb,
		logger: logger,
	}
}

// cooldownKey namespaces the reservation. This Redis is shared with the
// refresh tokens and the CSRF session store, so the prefix is what keeps the
// three from ever reading each other's keys.
func cooldownKey(userID, contestProjectID int32) string {
	return fmt.Sprintf("submission_cooldown:%d:%d", userID, contestProjectID)
}

func (c *SubmissionCooldown) Reserve(userID, contestProjectID int32, cooldownSeconds int32) (bool, int32, error) {
	if cooldownSeconds <= 0 {
		return true, 0, nil
	}

	ctx := context.Background()
	key := cooldownKey(userID, contestProjectID)
	window := time.Duration(cooldownSeconds) * time.Second

	// SET NX is the whole mechanism: it both claims the slot and reports
	// whether somebody got there first, with no gap in between for a second
	// request to slip through.
	claimed, err := c.rdb.SetNX(ctx, key, 1, window).Result()
	if err != nil {
		c.logger.Errorf("Failed to reserve submission cooldown", "userId", userID, "contestProjectId", contestProjectID, "error", err)
		return false, 0, err
	}

	if claimed {
		return true, 0, nil
	}

	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		c.logger.Errorf("Failed to read submission cooldown TTL", "userId", userID, "contestProjectId", contestProjectID, "error", err)
		return false, 0, err
	}

	// The key can expire between the SET and the TTL, and a key with no
	// expiry should not exist at all. Either way there is nothing left to
	// wait for, so admit rather than reject with a nonsense countdown.
	if ttl <= 0 {
		c.logger.Warnf("Submission cooldown key had no remaining TTL; admitting", "userId", userID, "contestProjectId", contestProjectID)
		return true, 0, nil
	}

	// Rounded up: a sub-second remainder must not be reported as 0.
	remaining := int32((ttl + time.Second - 1) / time.Second)
	c.logger.Infof("Submission rejected by cooldown", "userId", userID, "contestProjectId", contestProjectID, "retryAfterSeconds", remaining)
	return false, remaining, nil
}
