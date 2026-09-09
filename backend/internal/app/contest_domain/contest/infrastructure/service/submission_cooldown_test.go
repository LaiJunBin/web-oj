package service_test

import (
	"testing"
	"time"

	"judge-backend/internal/app/contest_domain/contest/infrastructure/service"
	"judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/go-redis/redismock/v9"
)

const key = "submission_cooldown:7:3"

func newCooldown() (interface {
	Reserve(userID, contestProjectID int32, cooldownSeconds int32) (bool, int32, error)
}, redismock.ClientMock) {
	rdb, mock := redismock.NewClientMock()
	return service.NewSubmissionCooldown(rdb, logger.NewMockLogger()), mock
}

func TestReserveAdmitsWhenTheSlotIsFree(t *testing.T) {
	cooldown, mock := newCooldown()
	mock.ExpectSetNX(key, 1, 30*time.Second).SetVal(true)

	ok, retryAfter, err := cooldown.Reserve(7, 3, 30)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !ok {
		t.Error("expected the first submission to be admitted")
	}
	if retryAfter != 0 {
		t.Errorf("retryAfter = %d, want 0", retryAfter)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected all mock expectations to be met, got %v", err)
	}
}

func TestReserveRejectsWhileTheSlotIsHeld(t *testing.T) {
	cooldown, mock := newCooldown()
	mock.ExpectSetNX(key, 1, 30*time.Second).SetVal(false)
	mock.ExpectTTL(key).SetVal(20 * time.Second)

	ok, retryAfter, err := cooldown.Reserve(7, 3, 30)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if ok {
		t.Error("expected a second submission inside the window to be rejected")
	}
	if retryAfter != 20 {
		t.Errorf("retryAfter = %d, want 20", retryAfter)
	}
}

// A client told to wait 0 seconds retries immediately and is rejected again,
// so a sub-second remainder has to round up.
func TestReserveRoundsUpASubSecondRemainder(t *testing.T) {
	cooldown, mock := newCooldown()
	mock.ExpectSetNX(key, 1, 30*time.Second).SetVal(false)
	mock.ExpectTTL(key).SetVal(200 * time.Millisecond)

	_, retryAfter, err := cooldown.Reserve(7, 3, 30)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if retryAfter != 1 {
		t.Errorf("retryAfter = %d, want 1", retryAfter)
	}
}

// The key can expire between the SET and the TTL. There is nothing left to
// wait for, so the submission goes through rather than being refused with a
// countdown of zero.
func TestReserveAdmitsWhenTheKeyVanishedBeforeTheTTLRead(t *testing.T) {
	cooldown, mock := newCooldown()
	mock.ExpectSetNX(key, 1, 30*time.Second).SetVal(false)
	mock.ExpectTTL(key).SetVal(-2 * time.Second)

	ok, _, err := cooldown.Reserve(7, 3, 30)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !ok {
		t.Error("expected admission once the reservation had already expired")
	}
}

// A contest running without a cooldown must not touch Redis at all — the
// mock asserting no calls is the point of this test.
func TestReserveIsANoOpWhenTheCooldownIsZero(t *testing.T) {
	cooldown, mock := newCooldown()

	ok, retryAfter, err := cooldown.Reserve(7, 3, 0)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !ok || retryAfter != 0 {
		t.Errorf("ok = %v, retryAfter = %d; want true, 0", ok, retryAfter)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected no redis calls, got %v", err)
	}
}
