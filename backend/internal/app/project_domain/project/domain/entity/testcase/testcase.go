package testcase

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
)

var (
	ErrorDependSelf = errors.New("cannot depend on self")
)

type Testcase struct {
	id            int32
	gradingItemID int32
	name          Name
	command       string
	score         float64
	isSample      bool
	depends  []int32

	sharedtypes.TimeStamps
}

type TestcaseConfig struct {
	GradingItemID int32
	Name          string
	Command       string
	Score         float64
	IsSample      bool
	Depends       []int32
}

func New(config *TestcaseConfig) (*Testcase, error) {
	name, err := NewName(config.Name)
	if err != nil {
		return nil, err
	}

	return &Testcase{
		gradingItemID: config.GradingItemID,
		name:          name,
		command:       config.Command,
		score:         config.Score,
		isSample:      config.IsSample,
		depends: config.Depends,
	}, nil
}

func (t *Testcase) ID() int32 {
	return t.id
}

func (t *Testcase) GradingItemID() int32 {
	return t.gradingItemID
}

func (t *Testcase) Name() Name {
	return t.name
}

func (t *Testcase) Command() string {
	return t.command
}

func (t *Testcase) Score() float64 {
	return t.score
}

func (t *Testcase) IsSample() bool {
	return t.isSample
}

func (t *Testcase) Depends() []int32 {
	return t.depends
}

func (t *Testcase) SetID(id int32) {
	t.id = id
}

func (t *Testcase) SetGradingItemID(gradingItemID int32) {
	t.gradingItemID = gradingItemID
}

func (t *Testcase) SetName(name string) error {
	n, err := NewName(name)
	if err != nil {
		return err
	}

	t.name = n
	return nil
}

func (t *Testcase) SetCommand(command string) {
	t.command = command
}

func (t *Testcase) SetScore(score float64) {
	t.score = score
}

func (t *Testcase) SetIsSample(isSample bool) {
	t.isSample = isSample
}

func (t *Testcase) SetDepends(depends []int32) error {
	// if depends has self id, throw error
	for _, id := range depends {
		if id == t.id {
			return ErrorDependSelf
		}
	}

	t.depends = depends
	return nil
}
