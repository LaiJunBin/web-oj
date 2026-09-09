package jsonpresenter_test

import (
	"encoding/json"
	"testing"
	"time"

	"judge-backend/api"
	jsonpresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/json/contestant"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	listuserprojectsubmissions "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_submissions"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func listEntries(t *testing.T, subs ...*dto.Submission) []map[string]any {
	t.Helper()
	p := jsonpresenter.NewListUserProjectSubmissionsPresenter()
	resp, ok := p.Output(&listuserprojectsubmissions.Output{Submissions: subs}).(api.ListUserProjectSubmissions200JSONResponse)
	require.True(t, ok, "unexpected response type %T", resp)

	// Round-trip through JSON: the contract is about which keys appear on
	// the wire, and omitempty only shows there.
	raw, err := json.Marshal(resp.Data)
	require.NoError(t, err)
	var entries []map[string]any
	require.NoError(t, json.Unmarshal(raw, &entries))
	return entries
}

// The list carries the staff note exactly as the detail does: present when
// written, absent (not null, not "") when there is none — and independent
// of reason, which may be present or absent on its own.
func TestListUserProjectSubmissions_StaffNoteTwoStates(t *testing.T) {
	now := time.Now()
	entries := listEntries(t,
		&dto.Submission{ID: "with-note", Status: "Judged", Score: 100, TotalScore: 100, CreatedAt: now,
			StaffNote: strPtr("做得好，注意第 2 題的邊界")},
		&dto.Submission{ID: "without-note", Status: "Queued", CreatedAt: now},
		&dto.Submission{ID: "error-with-both", Status: "Error", CreatedAt: now,
			Reason: strPtr("評測前置步驟失敗，請聯絡現場工作人員"), StaffNote: strPtr("已通知助教處理，請稍候再提交")},
	)
	require.Len(t, entries, 3)

	withNote, withoutNote, errorWithBoth := entries[0], entries[1], entries[2]

	assert.Equal(t, "做得好，注意第 2 題的邊界", withNote["staffNote"])
	_, hasReason := withNote["reason"]
	assert.False(t, hasReason, "a note on a healthy judgment must not invent a reason")

	_, hasNote := withoutNote["staffNote"]
	assert.False(t, hasNote, "an entry without a note must omit the field, not send null or \"\"")
	_, hasReason = withoutNote["reason"]
	assert.False(t, hasReason)

	assert.Equal(t, "評測前置步驟失敗，請聯絡現場工作人員", errorWithBoth["reason"])
	assert.Equal(t, "已通知助教處理，請稍候再提交", errorWithBoth["staffNote"])
}
