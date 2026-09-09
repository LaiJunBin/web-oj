package bundle

import (
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"fmt"
	"strings"
)

// Problem is one fault in an image bundle, located by its path in image.json.
type Problem struct {
	Path    string
	Message string
}

func (p Problem) Error() string { return p.Path + ": " + p.Message }

type Problems []Problem

func (p Problems) Error() string {
	parts := make([]string, len(p))
	for i, problem := range p {
		parts[i] = problem.Error()
	}
	return strings.Join(parts, "; ")
}

// Validate reports everything wrong with the metadata in one pass, for the
// same reason the project bundle does: a hand-written file should not need one
// upload per typo.
func (i *Image) Validate() Problems {
	var problems Problems
	add := func(path, format string, args ...any) {
		problems = append(problems, Problem{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	if i.FormatVersion <= 0 {
		add("formatVersion", "missing")
	} else if i.FormatVersion > FormatVersion {
		add("formatVersion", "此匯出檔是 %d 版，本系統只支援到 %d 版，請升級後再匯入", i.FormatVersion, FormatVersion)
	}

	if i.Kind != KindImage {
		add("kind", "必須是 %q，得到 %q", KindImage, i.Kind)
	}

	if strings.TrimSpace(i.Tag) == "" {
		add("tag", "不可為空")
	} else if imagetag.ContainsDeletedMarker(i.Tag) {
		// Reserved namespace of rename-on-delete: importing such a tag could
		// collide with a soft-deleted row's renamed tag. Rejected here so an
		// archive import gets a 400 problem list, not the factory's plainer
		// error deeper in the flow.
		add("tag", "不可包含保留字 %q（系統用於標記已刪除映像）", imagetag.DeletedTagMarker)
	}

	if i.Mode != ModeSource && i.Mode != ModeArchive {
		add("mode", "必須是 %q 或 %q，得到 %q", ModeSource, ModeArchive, i.Mode)
	}

	// A bundle without a fingerprint cannot be told apart from whatever the
	// receiving installation already has under the same tag, which is the one
	// question the importer has to answer before it can decide anything.
	if strings.TrimSpace(i.Fingerprint) == "" {
		add("fingerprint", "不可為空 —— 沒有它就無法判斷與系統既有的映像是否相同")
	}

	return problems
}

// PayloadFile is the entry the bundle's mode requires the zip to contain.
func (i *Image) PayloadFile() string {
	if i.Mode == ModeArchive {
		return ArchiveFile
	}
	return ContextFile
}
