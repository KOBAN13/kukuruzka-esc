package esc_core

type MutationPhase uint8

const (
	MutationIdle MutationPhase = iota
	MutationRunningSystem
	MutationApplyingCommands
)
