package ecs

import "fmt"

type ComponentSet map[ComponentID]struct{}

type AccessSet struct {
	Reads  ComponentSet
	Writes ComponentSet
}

type AccessConflict struct {
	Stage     StageID
	Component ComponentID
	First     string
	Second    string
}

func NewAccessSet() AccessSet {
	return AccessSet{
		Reads:  make(ComponentSet),
		Writes: make(ComponentSet),
	}
}

func (a AccessSet) Merge(other AccessSet) bool {
	var conflict = false

	for id := range other.Writes {
		if containsComponentSet(id, a.Writes) || containsComponentSet(id, other.Reads) {
			conflict = true
		}

		a.Writes[id] = struct{}{}
	}

	for id := range other.Reads {
		if containsComponentSet(id, a.Writes) {
			conflict = true
			continue
		}

		a.Reads[id] = struct{}{}
	}

	return conflict
}

func (a AccessSet) ConflictsWith(other AccessSet) []ComponentID {
	var conflict []ComponentID

	for id := range a.Writes {
		if containsComponentSet(id, other.Writes) || containsComponentSet(id, a.Reads) {
			conflict = append(conflict, id)
		}
	}

	for id := range other.Reads {
		if containsComponentSet(id, a.Writes) {
			conflict = append(conflict, id)
		}
	}

	return conflict
}

func (c AccessConflict) Error() string {
	return fmt.Sprintf(
		"%v: stage %d component %d: %s conflicts with %s",
		ErrAccessConflict,
		c.Stage,
		c.Component,
		c.First,
		c.Second,
	)
}
