package repository

import (
	"context"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// staffNoteInDB reads the raw field so the tests can tell "absent/null" from
// "empty string" — the API's clear must land as null, never as "".
func staffNoteInDB(t *testing.T, r *JudgmentRepository, id string) *string {
	t.Helper()
	var doc struct {
		StaffNote *string `bson:"staffNote"`
	}
	err := r.mdb.Database("judge").Collection("judgments").
		FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc)
	require.NoError(t, err)
	return doc.StaffNote
}

func TestUpdateStaffNote_WritesAndReturnsThePostUpdateEntity(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-staff-note-set"
	seedJudgment(t, r, id, judgestatus.JUDGING)

	note := "第 3 題的輸出多了一個換行"
	updated, err := r.UpdateStaffNote(id, &note)

	require.NoError(t, err)
	require.NotNil(t, updated)
	require.NotNil(t, updated.StaffNote())
	assert.Equal(t, note, *updated.StaffNote())
	// The whole document comes back, not just the field the update touched:
	// the handler publishes judgeupdated from this entity.
	assert.EqualValues(t, 1, updated.ContestProjectID())
	assert.Len(t, updated.GradingItems(), 1)
	assert.True(t, updated.Status().IsJudging(), "the note must not disturb the judgment's status")

	stored := staffNoteInDB(t, r, id)
	require.NotNil(t, stored)
	assert.Equal(t, note, *stored)
}

// Last write wins, verbatim: no merge, no history.
func TestUpdateStaffNote_LastWriteWins(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-staff-note-overwrite"
	seedJudgment(t, r, id, judgestatus.JUDGED)

	first, second := "first", "second"
	_, err := r.UpdateStaffNote(id, &first)
	require.NoError(t, err)
	updated, err := r.UpdateStaffNote(id, &second)
	require.NoError(t, err)

	require.NotNil(t, updated.StaffNote())
	assert.Equal(t, "second", *updated.StaffNote())
	stored := staffNoteInDB(t, r, id)
	require.NotNil(t, stored)
	assert.Equal(t, "second", *stored)
}

// A nil note clears: the stored value becomes null and the returned entity
// carries no note, so the published frame omits the field and the client
// drops what it was showing.
func TestUpdateStaffNote_NilClears(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-staff-note-clear"
	seedJudgment(t, r, id, judgestatus.JUDGED)

	note := "to be retracted"
	_, err := r.UpdateStaffNote(id, &note)
	require.NoError(t, err)

	updated, err := r.UpdateStaffNote(id, nil)

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Nil(t, updated.StaffNote())
	assert.Nil(t, staffNoteInDB(t, r, id))
}

// The note is orthogonal to the judgment's lifecycle: writing it while a
// worker owns the document must leave every other field alone, and the
// worker's later conditional saves must still match.
func TestUpdateStaffNote_TouchesOnlyTheNote(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-staff-note-isolated"
	seedJudgment(t, r, id, judgestatus.JUDGING)

	note := "note while judging"
	_, err := r.UpdateStaffNote(id, &note)
	require.NoError(t, err)

	assert.Equal(t, "Judging", statusInDB(t, r, id))
	// The worker's cancel path still finds a Judging document to transition.
	cancelled, err := r.TryCancel(id, "cancelled after note", time.Now())
	require.NoError(t, err)
	require.NotNil(t, cancelled, "the note write must not have consumed the Judging status the cancel filter matches on")
	require.NotNil(t, cancelled.StaffNote())
	assert.Equal(t, note, *cancelled.StaffNote(), "the note survives the status transition")
}

// An unknown id is (nil, nil): the API checked existence already, so the
// judgment vanished in between — nothing to publish, nothing to retry.
func TestUpdateStaffNote_UnknownIDIsNilNil(t *testing.T) {
	r := newRepoForTest(t)

	note := "nobody home"
	updated, err := r.UpdateStaffNote("judgment-staff-note-missing", &note)

	assert.NoError(t, err)
	assert.Nil(t, updated)
}

// FindByID's explicit projection must include the note, or every re-read
// path (and the judgeupdated built from it) would silently drop it.
func TestFindByID_PreservesStaffNote(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-staff-note-findbyid"
	seedJudgment(t, r, id, judgestatus.JUDGED)
	note := "projected"
	_, err := r.UpdateStaffNote(id, &note)
	require.NoError(t, err)

	found, err := r.FindByID(id)

	require.NoError(t, err)
	require.NotNil(t, found.StaffNote())
	assert.Equal(t, note, *found.StaffNote())
}
