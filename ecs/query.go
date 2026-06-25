package ecs

import (
	"fmt"
	"slices"
)

type QueryDescriptor struct {
	Name    string
	With    []ComponentID
	Without []ComponentID
	Reads   []ComponentID
	Writes  []ComponentID
}

type QueryBuilder struct {
	world      *World
	name       string
	with       []ComponentID
	without    []ComponentID
	reads      []ComponentID
	writes     []ComponentID
	buildError error
}

type Query struct {
	world       *World
	name        string
	descriptor  QueryDescriptor
	matches     []queryArchetypePlan
	seenVersion uint64
	access      AccessSet
}

type queryArchetypePlan struct {
	archetype *archetype
}

func NewQuery(world *World, name string) *QueryBuilder {
	return &QueryBuilder{
		world:      world,
		name:       name,
		with:       []ComponentID{},
		without:    []ComponentID{},
		reads:      []ComponentID{},
		writes:     []ComponentID{},
		buildError: nil,
	}
}

func (builder *QueryBuilder) With(component ComponentToken) *QueryBuilder {
	if builder.buildError != nil {
		return builder
	}

	var componentId, err = builder.world.registry.ID(component)

	if err != nil {
		builder.buildError = err
		return builder
	}

	if containsComponent(componentId, builder.with) {
		return builder
	}

	builder.with = append(builder.with, componentId)

	return builder
}

func (builder *QueryBuilder) Without(component ComponentToken) *QueryBuilder {
	if builder.buildError != nil {
		return builder
	}

	var componentId, err = builder.world.registry.ID(component)

	if err != nil {
		builder.buildError = err
		return builder
	}

	if containsComponent(componentId, builder.with) {
		builder.buildError = fmt.Errorf("%w: component cannot be both With and Without", ErrQueryAccess)
	}

	if containsComponent(componentId, builder.without) {
		return builder
	}

	builder.without = append(builder.without, componentId)

	return builder
}

func (builder *QueryBuilder) Read(component ComponentToken) *QueryBuilder {
	if builder.buildError != nil {
		return builder
	}

	var componentId, err = builder.world.registry.ID(component)

	if err != nil {
		builder.buildError = err
		return builder
	}

	if containsComponent(componentId, builder.without) {
		builder.buildError = fmt.Errorf("%w: component cannot be both Read and Without", ErrQueryAccess)
		return builder
	}

	if containsComponent(componentId, builder.writes) {
		builder.buildError = fmt.Errorf("%w: component cannot be both Read and Write", ErrQueryAccess)
		return builder
	}

	if containsComponent(componentId, builder.with) {
		builder.with = append(builder.with, componentId)
	}

	if containsComponent(componentId, builder.reads) {
		builder.reads = append(builder.reads, componentId)
	}

	return builder
}

func (builder *QueryBuilder) Write(component ComponentToken) *QueryBuilder {
	if builder.buildError != nil {
		return builder
	}

	var componentId, err = builder.world.registry.ID(component)

	if err != nil {
		builder.buildError = err
		return builder
	}

	if containsComponent(componentId, builder.without) {
		builder.buildError = fmt.Errorf("%w: component cannot be both Read and Without", ErrQueryAccess)
		return builder
	}

	if containsComponent(componentId, builder.reads) {
		builder.buildError = fmt.Errorf("%w: component cannot be both Read and Write", ErrQueryAccess)
		return builder
	}

	if containsComponent(componentId, builder.with) {
		builder.with = append(builder.with, componentId)
	}

	if containsComponent(componentId, builder.writes) {
		builder.reads = append(builder.reads, componentId)
	}
	return builder
}

func (builder *QueryBuilder) Build() (*Query, error) {
	if builder.buildError != nil {
		return nil, builder.buildError
	}

	return &Query{
		world: builder.world,
		name:  builder.name,
		descriptor: QueryDescriptor{
			Name:    builder.name,
			With:    append([]ComponentID{}, builder.with...),
			Without: append([]ComponentID{}, builder.without...),
			Reads:   append([]ComponentID{}, builder.reads...),
			Writes:  append([]ComponentID{}, builder.writes...),
		},
		access: NewAccessSet(),
	}, nil
}

func containsComponent(id ComponentID, ids []ComponentID) bool {
	_, ok := slices.BinarySearch(ids, id)
	return ok
}

func containsComponentAny(id any, ids []any) error {
	var componentId, ok = id.(ComponentID)

	if !ok {
		return fmt.Errorf("%w: got %T, want %T", ErrInvalidComponentID, id, ComponentID(0))
	}

	for _, raw := range ids {
		var item, ok = raw.(ComponentID)
		if !ok {
			return fmt.Errorf("%w: ids contains %T, want %T", ErrInvalidComponentID, raw, ComponentID(0))
		}

		if item == componentId {
			return nil
		}
	}

	return nil
}

func containsComponentSet(id ComponentID, ids ComponentSet) bool {
	_, ok := ids[id]

	return ok
}
