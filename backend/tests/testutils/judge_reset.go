package testutils

import (
	"fmt"
	"net/http"
	"time"

	"judge-backend/internal/config"
	"judge-backend/tests/setup"
)

// ResetJudgeState drives the app's judge test seam: drain in-flight event
// handlers, empty every queue, wait for the workers to go idle, and restart any
// worker a previous scenario stopped. The BDD scenarios run in a separate
// process from the app, so none of that can be done in memory.
//
// Every initializer whose scenarios can reach the judge pipeline needs this,
// and needs it BEFORE the database resets: a judgment left over from the
// previous scenario would otherwise be picked up mid-reset and write ghost
// documents into the freshly cleaned MongoDB.
//
// Connection errors are retried: the suite has no explicit server-readiness
// gate, so the very first scenario may run before the seam listener is up.
func ResetJudgeState(cfg *config.Config) error {
	url := fmt.Sprintf("http://localhost:%d/judge/reset", cfg.Test.Server.Port+setup.JudgeSeamPortOffset)
	deadline := time.Now().Add(30 * time.Second)
	for {
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			if time.Now().Before(deadline) {
				time.Sleep(250 * time.Millisecond)
				continue
			}
			return fmt.Errorf("judge reset seam unreachable: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("judge reset seam returned %d (workers still busy?)", resp.StatusCode)
		}
		return nil
	}
}
