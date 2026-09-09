package provider

// JudgePriorityStrategyValidator validates whether a strategy name is registered in the judge domain.
// Implemented by judge_domain's priority strategy provider, bridged at the composition root.
type JudgePriorityStrategyValidator interface {
	HasStrategy(strategyName string) bool
}
