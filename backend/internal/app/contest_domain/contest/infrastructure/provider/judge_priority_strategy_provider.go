package provider

import (
	domainprovider "judge-backend/internal/app/contest_domain/contest/domain/provider"
)

// JudgeAvailableStrategiesProvider is defined at point of use to avoid cross-domain import.
// The composition root bridges judge_domain/prioritystrategies.Provider to this interface.
type JudgeAvailableStrategiesProvider interface {
	GetAvailableStrategies() []string
}

type JudgePriorityStrategyProvider struct {
	strategies map[string]struct{}
}

func NewJudgePriorityStrategyProvider(
	provider JudgeAvailableStrategiesProvider,
) domainprovider.JudgePriorityStrategyValidator {
	strategies := provider.GetAvailableStrategies()
	judgePriorityStrategyProvider := &JudgePriorityStrategyProvider{
		strategies: map[string]struct{}{},
	}

	for _, strategy := range strategies {
		judgePriorityStrategyProvider.strategies[strategy] = struct{}{}
	}

	return judgePriorityStrategyProvider
}

func (p *JudgePriorityStrategyProvider) GetStrategies() []string {
	strategies := make([]string, 0, len(p.strategies))
	for name := range p.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

func (p *JudgePriorityStrategyProvider) HasStrategy(strategyName string) bool {
	_, ok := p.strategies[strategyName]
	return ok
}
