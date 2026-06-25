package esc_core

import (
	"slices"
	"strconv"
	"strings"
)

type componentSignature struct {
	ids []ComponentID
	key string
}

func newComponentSignature(ids []ComponentID) *componentSignature {
	copied := append([]ComponentID(nil), ids...)

	slices.Sort(copied)

	copied = slices.Compact(copied)

	var parts = make([]string, len(copied))

	for i, id := range copied {
		parts[i] = strconv.FormatUint(uint64(id), 10)
	}

	return &componentSignature{
		ids: copied,
		key: strings.Join(parts, ","),
	}
}

func (component *componentSignature) containsAll(ids []ComponentID) bool {
	for _, id := range ids {
		if _, ok := slices.BinarySearch(component.ids, id); !ok {
			return false
		}
	}

	return true
}

func (component *componentSignature) containsComponent(id ComponentID) bool {
	_, ok := slices.BinarySearch(component.ids, id)
	return ok
}

func (component *componentSignature) excludesAll(ids []ComponentID) bool {
	for _, id := range ids {
		if _, ok := slices.BinarySearch(component.ids, id); ok {
			return false
		}
	}

	return true
}
