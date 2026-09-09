package handler

import (
	"testing"

	"judge-backend/internal/app/contest_domain/contest/mocks"
	"judge-backend/internal/app/contest_domain/events/submissionjudged"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
)

// TestJudgedHandler_WiringPin_ClaimProtocolAndPrefixedKey pins the
// REGISTRATION (same template as the judge domain's
// TestCancelJudge_WiringPin_ClaimProtocolAndKey): RegisterJudgedHandler must
// subscribe through the claimed protocol, and its claim key must carry the
// "judged-contest-" prefix. The prefix is load-bearing: the event ID is the
// raw judgment ID — the exact keyspace the judge domain's judge_cancelled
// claims use — so an unprefixed key here would let one bounded context's
// claim swallow the other's delivery. Publishing on a real bus with the
// claim DENIED proves both wiring and key without running the handler, so
// the project query stays a bare strict mock; a registration downgraded to a
// raw SubscribeAsync, or a dropped prefix, fails this test.
func TestJudgedHandler_WiringPin_ClaimProtocolAndPrefixedKey(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	eventRepo := sharedmocks.NewEventRepository(t)
	eventRepo.On("TryClaimEvent", "judged-contest-judgment-1").Return(false, nil).Once()

	RegisterJudgedHandler(eb, eventRepo, mocks.NewProjectQuery(t), logger)

	eb.Publish(submissionjudged.New("judgment-1", 42))
	eb.WaitAsync()
}
