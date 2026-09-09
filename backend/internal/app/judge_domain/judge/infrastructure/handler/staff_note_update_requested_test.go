package handler

import (
	"errors"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/events/staffnoteupdaterequested"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newStaffNoteHandler(t *testing.T) (*StaffNoteUpdateRequested, *sharedmocks.EventBus, *judgemocks.JudgmentRepository) {
	eventbus := sharedmocks.NewEventBus(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	return &StaffNoteUpdateRequested{
		eventbus:     eventbus,
		judgmentRepo: judgmentRepo,
		logger:       sharedlogger.NewMockLogger(),
	}, eventbus, judgmentRepo
}

func notedJudgment(id string, note *string) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project:   judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		JudgeType: judgetype.JUDGE,
		CreatedAt: time.Now(),
	})
	j.SetID(id)
	j.SetStatus(judgestatus.JUDGED)
	j.SetStaffNote(note)
	return j
}

func strPtr(s string) *string { return &s }

// A repo failure is the retryable case: return the error so the claimed
// subscription retries, and publish nothing — a judgeupdated without the
// write behind it would show the contestant a note that was never saved.
func TestUpdateStaffNote_RepoError_ReturnsErrorNoPublish(t *testing.T) {
	h, _, judgmentRepo := newStaffNoteHandler(t)
	e := staffnoteupdaterequested.New("evt-1", "judgment-1", strPtr("看看第 3 題的輸出格式"))

	judgmentRepo.On("UpdateStaffNote", "judgment-1", strPtr("看看第 3 題的輸出格式")).Return(nil, errors.New("mongo unavailable")).Once()

	assert.Error(t, h.UpdateStaffNote(e))
}

// The judgment vanished between the API's existence check and the write:
// nothing to publish, and a retry cannot bring it back, so nil.
func TestUpdateStaffNote_NotFound_NilNoPublish(t *testing.T) {
	h, _, judgmentRepo := newStaffNoteHandler(t)
	e := staffnoteupdaterequested.New("evt-1", "judgment-1", strPtr("note"))

	judgmentRepo.On("UpdateStaffNote", "judgment-1", strPtr("note")).Return(nil, nil).Once()

	assert.NoError(t, h.UpdateStaffNote(e))
}

// A successful write publishes the SAME judgeupdated the judge pipeline
// emits, carrying the post-update note — this is how an open contestant
// stream learns about it without a dedicated event.
func TestUpdateStaffNote_Applied_PublishesJudgeUpdatedWithNote(t *testing.T) {
	h, eventbus, judgmentRepo := newStaffNoteHandler(t)
	e := staffnoteupdaterequested.New("evt-1", "judgment-1", strPtr("記得補上 README"))

	judgmentRepo.On("UpdateStaffNote", "judgment-1", strPtr("記得補上 README")).
		Return(notedJudgment("judgment-1", strPtr("記得補上 README")), nil).Once()
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgeupdated.JudgeUpdatedEvent) bool {
		d := evt.Data()
		return d.JudgmentID == "judgment-1" && d.StaffNote != nil && *d.StaffNote == "記得補上 README"
	})).Once()

	assert.NoError(t, h.UpdateStaffNote(e))
}

// Clearing travels the same path with a nil note, and the published frame
// must carry nil too — that absence is what tells the client to drop the
// note it was showing.
func TestUpdateStaffNote_Clear_PublishesJudgeUpdatedWithoutNote(t *testing.T) {
	h, eventbus, judgmentRepo := newStaffNoteHandler(t)
	e := staffnoteupdaterequested.New("evt-1", "judgment-1", nil)

	judgmentRepo.On("UpdateStaffNote", "judgment-1", (*string)(nil)).
		Return(notedJudgment("judgment-1", nil), nil).Once()
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgeupdated.JudgeUpdatedEvent) bool {
		d := evt.Data()
		return d.JudgmentID == "judgment-1" && d.StaffNote == nil
	})).Once()

	assert.NoError(t, h.UpdateStaffNote(e))
}
