package ecs

import "fmt"

type Entity struct {
	index      uint32
	generation uint32
}

func (e Entity) Index() uint32 {
	return e.index
}

func (e Entity) Generation() uint32 {
	return e.generation
}

func (e Entity) IsZero() bool {
	return e.index == 0 && e.generation == 0
}

func (e Entity) String() string {
	return fmt.Sprintf("Entity{%d:%d}", e.index, e.generation)
}

type entitySlot struct {
	generation uint32
	alive      bool
	location   entityLocation
}

type entityLocation struct {
	archetype *archetype
	row       int
}
