package esc_core

type StageID uint16

type Context struct {
	World        *World
	Commands     *CommandBuffer
	Resources    *Resources
	Tick         uint64
	DeltaSeconds float32
	Stage        StageID
}

type System interface {
	Name() string
	Stage() StageID
	Update(ctx *Context) error
	Access() AccessSet
	DebugQueries() []QueryDebugInfo
}
