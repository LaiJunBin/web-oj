package prioritystrategies

import (
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/common"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/query"
)

type Strategy common.Prioritizable[*judgment.Judgment]

type Provider interface {
	GetAvailableStrategies() []string
	HasStrategy(strategyName string) bool
	GetStrategy(strategyName string) (Strategy, error)
	GetDefaultStrategy() Strategy
}

type provider struct {
	strategyMap map[string]Strategy
}

func NewProvider(
	query query.JudgmentQuery,
) Provider {
	provider := &provider{
		strategyMap: make(map[string]Strategy),
	}

	provider.register(NewDefault())
	provider.register(NewSubmitTime())
	provider.register(NewTimeSinceLast(query))
	return provider
}

func (p *provider) register(strategy Strategy) {
	p.strategyMap[strategy.Name()] = strategy
}

func (p *provider) GetAvailableStrategies() []string {
	strategies := make([]string, 0, len(p.strategyMap))
	for name := range p.strategyMap {
		strategies = append(strategies, name)
	}
	return strategies
}

func (p *provider) HasStrategy(strategyName string) bool {
	_, ok := p.strategyMap[strategyName]
	return ok
}

func (p *provider) GetStrategy(strategyName string) (Strategy, error) {
	strategy, ok := p.strategyMap[strategyName]
	if !ok {
		return nil, fmt.Errorf("strategy %s not found", strategyName)
	}
	return strategy, nil
}

func (p *provider) GetDefaultStrategy() Strategy {
	return p.strategyMap["default"]
}
