package service

import (
	"bytes"
	"fmt"
	"strings"

	"judge-backend/internal/app/contest_domain/activity/domain/service"

	"github.com/xuri/excelize/v2"
)

type ScoreDetailReportGenerator struct{}

func NewScoreDetailReportGenerator() service.ScoreDetailReportGenerator {
	return &ScoreDetailReportGenerator{}
}

var columnWidths = []float64{8, 45, 45, 10, 10, 12}

// Excel forbids these characters in SHEET NAMES only — cell content is
// unrestricted, which is why the contestant's real name goes untouched into
// row 1 while the tab gets this replacer.
var sheetNameReplacer = strings.NewReplacer(
	"[", "_", "]", "_", ":", "_", "*", "_", "?", "_", "/", "_", "\\", "_",
)

// sanitizeSheetName turns "Name(Username)" into a name Excel accepts: the
// forbidden characters become underscores, a leading/trailing apostrophe is
// trimmed (also forbidden), and the " #<userID>" suffix is reserved BEFORE
// the 31-rune cap is applied — user id is the only part guaranteed to tell
// two contestants apart once names have been sanitised and truncated, and
// excelize's NewSheet answers a DUPLICATE name with the existing sheet's
// index rather than an error, which used to merge truncated twins into one
// sheet. "History" is reserved by excelize and falls back like an empty name.
func sanitizeSheetName(name, username string, userID int32) string {
	base := sheetNameReplacer.Replace(fmt.Sprintf("%s(%s)", name, username))
	base = strings.Trim(base, "'")
	if base == "" || strings.EqualFold(base, "History") {
		base = "contestant"
	}

	suffix := fmt.Sprintf(" #%d", userID)
	if budget := 31 - len([]rune(suffix)); len([]rune(base)) > budget {
		base = string([]rune(base)[:budget])
	}
	return base + suffix
}

func (g *ScoreDetailReportGenerator) Generate(reports []*service.ContestantScoreReport) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	passedStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#C6EFCE"}, Pattern: 1},
		Font: &excelize.Font{Color: "#006100"},
	})
	if err != nil {
		return nil, err
	}

	failedStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFC7CE"}, Pattern: 1},
		Font: &excelize.Font{Color: "#9C0006"},
	})
	if err != nil {
		return nil, err
	}

	headers := []string{"項次", "評分項目", "補充說明", "得分", "滿分", "狀態"}

	// Every excelize error surfaces: this export is the contest's official
	// score record, and a swallowed error used to mean a silently missing or
	// half-written sheet behind a 200 download.
	for _, report := range reports {
		sheetName := sanitizeSheetName(report.Name, report.Username, report.UserID)
		if _, err := f.NewSheet(sheetName); err != nil {
			return nil, fmt.Errorf("create sheet %q: %w", sheetName, err)
		}

		cols := []string{"A", "B", "C", "D", "E", "F"}
		for i, col := range cols {
			if err := f.SetColWidth(sheetName, col, col, columnWidths[i]); err != nil {
				return nil, fmt.Errorf("set column width on %q: %w", sheetName, err)
			}
		}

		// Row 1 carries the contestant's REAL name — the tab above had to be
		// sanitised, the cell does not.
		if err := f.SetCellValue(sheetName, "A1", fmt.Sprintf("%s(%s)", report.Name, report.Username)); err != nil {
			return nil, fmt.Errorf("write title on %q: %w", sheetName, err)
		}

		setCell := func(col, row int, v interface{}) (string, error) {
			cell, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				return "", fmt.Errorf("resolve cell (%d,%d) on %q: %w", col, row, sheetName, err)
			}
			if err := f.SetCellValue(sheetName, cell, v); err != nil {
				return "", fmt.Errorf("write %s on %q: %w", cell, sheetName, err)
			}
			return cell, nil
		}

		// Headers on row 2, data from row 3.
		for col, header := range headers {
			if _, err := setCell(col+1, 2, header); err != nil {
				return nil, err
			}
		}

		for row, item := range report.Items {
			rowNum := row + 3
			if _, err := setCell(1, rowNum, item.Label); err != nil {
				return nil, err
			}
			if _, err := setCell(2, rowNum, item.Name); err != nil {
				return nil, err
			}
			if _, err := setCell(3, rowNum, item.Description); err != nil {
				return nil, err
			}
			if _, err := setCell(4, rowNum, item.Score); err != nil {
				return nil, err
			}
			if _, err := setCell(5, rowNum, item.TotalScore); err != nil {
				return nil, err
			}

			statusCell, err := setCell(6, rowNum, item.Status)
			if err != nil {
				return nil, err
			}
			style := 0
			switch item.Status {
			case "Passed":
				style = passedStyle
			case "Failed":
				style = failedStyle
			}
			if style != 0 {
				if err := f.SetCellStyle(sheetName, statusCell, statusCell, style); err != nil {
					return nil, fmt.Errorf("style %s on %q: %w", statusCell, sheetName, err)
				}
			}
		}
	}

	// Drop the default Sheet1 — unless it is the only sheet (an export with
	// zero contestants), which excelize refuses to leave sheetless.
	if len(reports) > 0 {
		if err := f.DeleteSheet("Sheet1"); err != nil {
			return nil, fmt.Errorf("remove the default sheet: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
