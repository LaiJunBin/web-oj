package streamusersubmission

import (
	"errors"
	"testing"
	"time"

	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
)

// The resolver is the stream's eventual-consistency contract: at most TTL of
// staleness after a visibility change, at most one query per TTL per stream
// no matter how dense the event burst, and a failed refresh degrades to the
// last known set instead of dropping frames or failing open.
func TestNewSampleSetResolver(t *testing.T) {
	start := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

	t.Run("serves the seed without fetching until the TTL passes", func(t *testing.T) {
		clock := start
		fetches := 0
		resolve := newSampleSetResolver([]int32{1, 3}, func() ([]int32, error) {
			fetches++
			return []int32{2}, nil
		}, 5*time.Second, func() time.Time { return clock }, sharedlogger.NewMockLogger())

		assert.Equal(t, map[int32]bool{1: true, 3: true}, resolve())
		clock = start.Add(5 * time.Second) // exactly at the TTL: still fresh
		assert.Equal(t, map[int32]bool{1: true, 3: true}, resolve())
		assert.Zero(t, fetches)
	})

	t.Run("refreshes once past the TTL and memoises the new set", func(t *testing.T) {
		clock := start
		fetches := 0
		resolve := newSampleSetResolver([]int32{1}, func() ([]int32, error) {
			fetches++
			return []int32{2}, nil
		}, 5*time.Second, func() time.Time { return clock }, sharedlogger.NewMockLogger())

		clock = start.Add(6 * time.Second)
		assert.Equal(t, map[int32]bool{2: true}, resolve())
		assert.Equal(t, map[int32]bool{2: true}, resolve()) // within the new TTL
		assert.Equal(t, 1, fetches, "an event burst inside one TTL costs at most one query")
	})

	t.Run("a failed refresh keeps the last known set and retries next call", func(t *testing.T) {
		clock := start
		fetches := 0
		fail := true
		resolve := newSampleSetResolver([]int32{1}, func() ([]int32, error) {
			fetches++
			if fail {
				return nil, errors.New("db down")
			}
			return []int32{9}, nil
		}, 5*time.Second, func() time.Time { return clock }, sharedlogger.NewMockLogger())

		clock = start.Add(6 * time.Second)
		assert.Equal(t, map[int32]bool{1: true}, resolve(), "failure must not widen or drop the filter")
		fail = false
		assert.Equal(t, map[int32]bool{9: true}, resolve(), "the failed refresh must not extend the TTL")
		assert.Equal(t, 2, fetches)
	})
}
