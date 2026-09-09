package service

import (
	"fmt"
	"time"

	importcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/import_contest"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type projectLibrary struct{ db *gorm.DB }

func NewProjectLibrary(db *gorm.DB) importcontest.ProjectLibrary {
	return &projectLibrary{db: db}
}

// FindIDByName returns 0 for both "no match" and "more than one match".
//
// Project names carry no uniqueness constraint, so a name really can identify
// two different problems. Picking either would produce a contest that looks
// correctly configured and grades the wrong thing, which nobody would notice
// until the sitting was under way — so ambiguity is reported as unresolved and
// left to a person.
func (l *projectLibrary) FindIDByName(name string) (int32, error) {
	var ids []int32
	if err := l.db.Model(&model.Project{}).
		Where("name = ?", name).
		Order("id ASC").
		Limit(2).
		Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) != 1 {
		return 0, nil
	}
	return ids[0], nil
}

type contestNamer struct{ db *gorm.DB }

func NewContestNamer(db *gorm.DB) importcontest.ContestNamer {
	return &contestNamer{db: db}
}

// UniqueName appends the import time when the name is taken, matching what the
// project import does. Contest names are not unique either, so this is a
// courtesy — but two identically named sittings in the admin list help nobody
// pick the right one.
func (n *contestNamer) UniqueName(wanted string) (string, error) {
	taken, err := n.exists(wanted)
	if err != nil {
		return "", err
	}
	if !taken {
		return wanted, nil
	}

	candidate := fmt.Sprintf("%s (匯入 %s)", wanted, time.Now().Format("2006-01-02 15:04"))
	for i := 2; ; i++ {
		taken, err := n.exists(candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s (匯入 %s)-%d", wanted, time.Now().Format("2006-01-02 15:04"), i)
	}
}

func (n *contestNamer) exists(name string) (bool, error) {
	var count int64
	if err := n.db.Model(&model.Contest{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
