package ecs

type MutationPhase uint8

const (
	MutationIdle MutationPhase = iota
	MutationRunningSystem
	MutationApplyingCommands
)
