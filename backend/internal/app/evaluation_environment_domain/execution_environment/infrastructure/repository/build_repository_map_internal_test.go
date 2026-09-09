package repository

import (
	"fmt"
	"sync"
	"testing"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
)

// TestBuildRepository_BuildMap_ConcurrentAccess pins the buildMap mutex:
// queue workers (FindLastBuild), the async cancel path
// (FindCurrentBuildByTag), and API goroutines (Save) touch the map
// concurrently — a plain map here is a crash-level data race, which `go
// test -race` flags the moment this test runs against an unsynchronized
// implementation. Only map-hitting paths are exercised (nil gorm handle):
// the mutex under test guards the map, not the SQL fallbacks. (Writers go
// through SavePendingExclusive/ReleasePending — the reservation contract's
// only write paths since the unconditional Save was removed.)
func TestBuildRepository_BuildMap_ConcurrentAccess(t *testing.T) {
	r := &BuildRepository{buildMap: map[string]*build.Build{}}

	const goroutines = 8
	const iterations = 200
	var wg sync.WaitGroup
	start := make(chan struct{})

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				tag := fmt.Sprintf("map-race-%d:latest", i%10)
				b := build.New(&build.BuildConfig{
					Tag:    imagetag.MustNew(tag),
					Status: buildstatus.PENDING,
					Batch:  batch.MustNew(1),
				})
				// Only map-backed operations: the find variants with SQL
				// fallbacks would dereference the nil gorm handle the moment
				// a concurrent ReleasePending empties the key.
				_ = r.SavePendingExclusive(b)
				_ = r.FindCurrentBuildByTag(tag)
				r.ReleasePending(b)
			}
		}(g)
	}
	close(start)
	wg.Wait()
}

// TestBuildRepository_UnregisterPending_IdentityGuard drives the reviewer's
// exact sequence — cancel build A, enqueue build B, then A's DEFERRED commit
// cleanup runs — and pins that the late cleanup cannot delete B: Commit's
// map removal (extracted as unregisterPending so it is testable without a
// SQL database) only applies while the entry still IS the committing build.
func TestBuildRepository_UnregisterPending_IdentityGuard(t *testing.T) {
	r := &BuildRepository{buildMap: map[string]*build.Build{}}

	newBuild := func(batchValue int) *build.Build {
		return build.New(&build.BuildConfig{
			Tag:    imagetag.MustNew("identity-guard:latest"),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(batchValue),
		})
	}

	a := newBuild(1)
	if err := r.SavePendingExclusive(a); err != nil {
		t.Fatalf("registering A: %v", err)
	}

	// Cancel A: its commit cleanup removes the entry (it is still A's).
	r.unregisterPending(a)
	if got := r.FindCurrentBuildByTag(a.Tag()); got != nil {
		t.Fatalf("expected no current build after A's cancel cleanup, got %+v", got)
	}

	// Enqueue B while A's executing goroutine is still unwinding.
	b := newBuild(2)
	if err := r.SavePendingExclusive(b); err != nil {
		t.Fatalf("registering B: %v", err)
	}

	// A's deferred commit finally runs its cleanup — it must NOT remove B.
	r.unregisterPending(a)

	if got := r.FindCurrentBuildByTag(b.Tag()); got != b {
		t.Fatalf("build A's late cleanup removed build B from the map — B can no longer be found or cancelled (got %+v)", got)
	}
}

// TestBuildRepository_SavePendingExclusive_RejectsWhileLive pins the atomic
// check-and-register: while a pending or building entry exists for the tag,
// a second registration is refused instead of overwriting it.
func TestBuildRepository_SavePendingExclusive_RejectsWhileLive(t *testing.T) {
	r := &BuildRepository{buildMap: map[string]*build.Build{}}

	a := build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew("exclusive:latest"),
		Status: buildstatus.PENDING,
		Batch:  batch.MustNew(1),
	})
	if err := r.SavePendingExclusive(a); err != nil {
		t.Fatalf("registering A: %v", err)
	}

	b := build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew("exclusive:latest"),
		Status: buildstatus.PENDING,
		Batch:  batch.MustNew(1),
	})
	if err := r.SavePendingExclusive(b); err != repository.ErrorBuildAlreadyInProgress {
		t.Fatalf("expected ErrorBuildAlreadyInProgress, got %v", err)
	}
	if got := r.FindCurrentBuildByTag(a.Tag()); got != a {
		t.Fatalf("the live pending build must not be overwritten, got %+v", got)
	}
}
