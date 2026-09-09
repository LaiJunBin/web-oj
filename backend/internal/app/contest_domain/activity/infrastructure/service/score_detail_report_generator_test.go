package service

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"judge-backend/internal/app/contest_domain/activity/domain/service"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// Excel forbids [ ] : * ? / \ in SHEET NAMES (cell content is unrestricted)
// and caps them at 31 runes. The generator used to hand excelize the raw
// "Name(Username)" and ignore every error, so a contestant whose name or
// username carried a forbidden character lost their whole sheet silently, and
// two names agreeing on the first 31 runes silently merged into one sheet.
// These tests read the produced workbook back and pin the sanitised layout.

func scoreReport(userID int32, name, username string, rows int) *service.ContestantScoreReport {
	items := make([]*service.ScoreReportItem, 0, rows)
	for i := 0; i < rows; i++ {
		items = append(items, &service.ScoreReportItem{
			Label:      fmt.Sprintf("%d", i+1),
			Name:       fmt.Sprintf("item-%d-of-%s", i+1, username),
			Score:      10,
			TotalScore: 10,
			Status:     "Passed",
		})
	}
	return &service.ContestantScoreReport{UserID: userID, Name: name, Username: username, Items: items}
}

func openWorkbook(t *testing.T, data []byte) *excelize.File {
	t.Helper()
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("generated workbook does not open: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func cellValue(t *testing.T, f *excelize.File, sheet, cell string) string {
	t.Helper()
	v, err := f.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("read %s!%s: %v", sheet, cell, err)
	}
	return v
}

func TestGenerate_SanitisesForbiddenSheetNameCharacters(t *testing.T) {
	g := NewScoreDetailReportGenerator()
	data, err := g.Generate([]*service.ContestantScoreReport{
		scoreReport(7, "王[小]明", "user:1", 2),
		scoreReport(8, "A/B\\C?D*E", "user2", 1),
	})
	assert.NoError(t, err)

	f := openWorkbook(t, data)
	sheets := f.GetSheetList()
	assert.Len(t, sheets, 2, "every contestant keeps a sheet, forbidden characters or not; got %v", sheets)

	assert.Contains(t, sheets, "王_小_明(user_1) #7")
	assert.Contains(t, sheets, "A_B_C_D_E(user2) #8")

	// The tab is sanitised; the REAL name lives untouched inside the sheet.
	assert.Equal(t, "王[小]明(user:1)", cellValue(t, f, "王_小_明(user_1) #7", "A1"))
	// Headers moved to row 2, data starts at row 3.
	assert.Equal(t, "項次", cellValue(t, f, "王_小_明(user_1) #7", "A2"))
	assert.Equal(t, "item-1-of-user:1", cellValue(t, f, "王_小_明(user_1) #7", "B3"))
	assert.Equal(t, "Passed", cellValue(t, f, "王_小_明(user_1) #7", "F3"))

	// The status cell keeps its fill style.
	styleID, err := f.GetCellStyle("王_小_明(user_1) #7", "F3")
	assert.NoError(t, err)
	assert.NotZero(t, styleID, "the Passed cell must keep its colour style")
}

func TestGenerate_TruncationCollisionKeepsBothSheets(t *testing.T) {
	long := strings.Repeat("甲", 29)
	g := NewScoreDetailReportGenerator()
	data, err := g.Generate([]*service.ContestantScoreReport{
		scoreReport(1, long, "userA", 3),
		scoreReport(2, long, "userB", 5),
	})
	assert.NoError(t, err)

	f := openWorkbook(t, data)
	sheets := f.GetSheetList()
	assert.Len(t, sheets, 2, "truncated twins must not merge; got %v", sheets)

	for _, report := range []struct {
		userID int32
		user   string
		rows   int
	}{{1, "userA", 3}, {2, "userB", 5}} {
		var sheet string
		for _, s := range sheets {
			if strings.HasSuffix(s, fmt.Sprintf("#%d", report.userID)) {
				sheet = s
			}
		}
		if assert.NotEmpty(t, sheet, "no sheet carries the #%d suffix", report.userID) {
			assert.LessOrEqual(t, len([]rune(sheet)), 31)
			assert.Equal(t, long+"("+report.user+")", cellValue(t, f, sheet, "A1"))
			assert.Equal(t, fmt.Sprintf("item-%d-of-%s", report.rows, report.user), cellValue(t, f, sheet, fmt.Sprintf("B%d", report.rows+2)))
			assert.Equal(t, "", cellValue(t, f, sheet, fmt.Sprintf("B%d", report.rows+3)), "the other contestant's rows must not bleed in")
		}
	}
}

func TestGenerate_NormalNamesKeepFullTab(t *testing.T) {
	g := NewScoreDetailReportGenerator()
	data, err := g.Generate([]*service.ContestantScoreReport{scoreReport(3, "Alice", "alice", 1)})
	assert.NoError(t, err)

	f := openWorkbook(t, data)
	assert.Equal(t, []string{"Alice(alice) #3"}, f.GetSheetList())
	assert.Equal(t, "Alice(alice)", cellValue(t, f, "Alice(alice) #3", "A1"))
}

// The sanitiser's edge branches are unreachable through Generate's composed
// "Name(Username)" shape (the parentheses keep it non-empty), so they are
// pinned directly.
func TestSanitizeSheetName_Edges(t *testing.T) {
	assert.Equal(t, "王_小_明(user_1) #7", sanitizeSheetName("王[小]明", "user:1", 7))
	// A leading apostrophe is forbidden by Excel and trimmed.
	assert.Equal(t, "n(u) #1", sanitizeSheetName("'n", "u", 1))
	// 31-rune cap holds with the suffix reserved first.
	long := sanitizeSheetName(strings.Repeat("x", 40), "user", 123)
	assert.LessOrEqual(t, len([]rune(long)), 31)
	assert.True(t, strings.HasSuffix(long, " #123"))
}
