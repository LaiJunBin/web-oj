package repository

import (
	"context"
	sharedtypes "judge-backend/internal/app/shared/types"
	"time"

	"github.com/redis/go-redis/v9"
)

type EventRepository struct {
	rdb    *redis.Client
	logger sharedtypes.Logger
}

func NewEventRepository(rdb *redis.Client, logger sharedtypes.Logger) sharedtypes.EventRepository {
	return &EventRepository{
		rdb:    rdb,
		logger: logger,
	}
}

func (r *EventRepository) TryClaimEvent(eventID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ok, err := r.rdb.SetNX(ctx, eventID, true, 5*time.Minute).Result()
	if err != nil {
		r.logger.Errorf("[EventRepository] Failed to claim event %s: %v", eventID, err)
		return false, err
	}
	return ok, nil
}

// ReleaseEvent removes the claim so the event can be re-attempted. The key
// format and TTL above stay untouched — a claim written by an older build is
// still releasable by this one and vice versa.
//
// The DEL is unconditional, and what makes that safe is that no claim can
// still be held by an attempt that has already lost it. The claim is taken
// and released within ONE handler invocation (each retry re-claims), so the
// only thing that matters is how long a single invocation runs — the retry
// count and its backoff are irrelevant here. Six of the seven claimed
// handlers are structurally bounded: each does an O(1) number of Mongo
// round-trips, every one of them under its own 10s context timeout, or no
// I/O at all. The exception is the cancel-rejudge handler: its member loop
// issues one TryCancel PER member with no loop-level deadline, so its worst
// case is N × 10s — around 30 members is already the whole 5-minute TTL,
// and a rejudge of that size ("re-run this whole task") is ordinary. The
// margin there is linear in the member count, not orders of magnitude.
//
// It stays acceptable because the handlers are idempotent by contract (an
// expired claim costs duplicate, harmless work), and reaching that worst
// case needs Mongo to time out on every member in turn. Giving that loop
// its own deadline would turn "unreachable in practice" into a structural
// guarantee, and is the cheap fix if this ever bites.
//
// Add a handler that can outlive the TTL and this stops being true: its
// claim expires mid-run, a redelivery claims the event afresh, and the
// original attempt's failure then deletes SOMEONE ELSE's claim. At that
// point the release must become ownership-checked — store a per-attempt
// token as the claim's value and compare-and-delete it (Lua) — which means
// changing both methods' signatures and threading the token through the
// claimed subscription and the three creation-style handlers.
func (r *EventRepository) ReleaseEvent(eventID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.rdb.Del(ctx, eventID).Err(); err != nil {
		r.logger.Errorf("[EventRepository] Failed to release event %s: %v", eventID, err)
		return err
	}
	return nil
}
