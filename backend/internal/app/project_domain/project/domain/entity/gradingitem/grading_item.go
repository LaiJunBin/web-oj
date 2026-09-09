package gradingitem

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem/outputformater"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem/scoringstrategy"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
)

var (
	ErrorInvalidScoringStrategy = errors.New("invalid scoring strategy")
	ErrorInvalidOutputFormat    = errors.New("invalid output format")
)

type GradingItem struct {
	id              int32
	projectID       project.ID
	name            Name
	description     string
	order           uint32
	score           float64
	scoringStrategy scoringstrategy.ScoringStrategy
	outputFormater  outputformater.OutputFormater
	dependTestcases []int32
	resultParser    string
	label           string
	sharedtypes.TimeStamps
}

type GradingItemConfig struct {
	ProjectID       project.ID
	Name            string
	Description     string
	Order           uint32
	Score           float64
	ScoringStrategy string
	OutputFormat    string
	DependTestcases []int32
	ResultParser    string
	Label           string
}

func New(config *GradingItemConfig) (*GradingItem, error) {
	name, err := NewName(config.Name)
	if err != nil {
		return nil, err
	}

	st, ok := scoringstrategy.Strategies[config.ScoringStrategy]
	if !ok {
		return nil, ErrorInvalidScoringStrategy
	}

	f, ok := outputformater.Formaters[config.OutputFormat]
	if !ok {
		return nil, ErrorInvalidOutputFormat
	}

	return &GradingItem{
		projectID:       config.ProjectID,
		name:            name,
		description:     config.Description,
		order:           config.Order,
		score:           config.Score,
		scoringStrategy: st,
		outputFormater:  f,
		dependTestcases: config.DependTestcases,
		resultParser:    config.ResultParser,
		label:           config.Label,
	}, nil
}

func (g *GradingItem) ID() int32 {
	return g.id
}

func (g *GradingItem) ProjectID() project.ID {
	return g.projectID
}

func (g *GradingItem) Name() Name {
	return g.name
}

func (g *GradingItem) Description() string {
	return g.description
}

func (g *GradingItem) Order() uint32 {
	return g.order
}

func (g *GradingItem) Score() float64 {
	return g.score
}

func (g *GradingItem) ScoringStrategy() scoringstrategy.ScoringStrategy {
	return g.scoringStrategy
}

func (g *GradingItem) OutputFormater() outputformater.OutputFormater {
	return g.outputFormater
}

func (g *GradingItem) DependTestcases() []int32 {
	return g.dependTestcases
}

func (g *GradingItem) SetID(id int32) {
	g.id = id
}

func (g *GradingItem) SetName(name string) error {
	n, err := NewName(name)
	if err != nil {
		return err
	}

	g.name = n
	return nil
}

func (g *GradingItem) SetDescription(description string) {
	g.description = description
}

func (g *GradingItem) SetOrder(order uint32) {
	g.order = order
}

func (g *GradingItem) SetScore(score float64) {
	g.score = score
}

func (g *GradingItem) SetScoringStrategy(scoringStrategy scoringstrategy.ScoringStrategy) {
	g.scoringStrategy = scoringStrategy
}

func (g *GradingItem) SetOutputFormater(outputFormater outputformater.OutputFormater) {
	g.outputFormater = outputFormater
}

func (g *GradingItem) SetDependTestcases(dependTestcases []int32) {
	g.dependTestcases = dependTestcases
}

func (g *GradingItem) ResultParser() string {
	return g.resultParser
}

func (g *GradingItem) SetResultParser(resultParser string) {
	g.resultParser = resultParser
}

func (g *GradingItem) Label() string {
	return g.label
}

func (g *GradingItem) SetLabel(label string) {
	g.label = label
}
