package esc_core

import (
	"fmt"
	"reflect"
)

type BundleBuilder struct {
	components []any
	seen       map[ComponentID]struct{}
	registry   *ComponentRegistry
	err        error
}

type Bundle interface {
	Apply(*BundleBuilder) error
}

type BundleFunc func(*BundleBuilder) error

func NewBundleBuilder() *BundleBuilder {
	return &BundleBuilder{
		components: make([]any, 0),
		seen:       make(map[ComponentID]struct{}),
	}
}

func (f BundleFunc) Apply(builder *BundleBuilder) error {
	return f(builder)
}

func (b *BundleBuilder) With(component any) error {
	if err := containsComponentAny(component, b.components); err != nil {
		return err
	}

	var token = ComponentToken{
		Type: reflect.TypeOf(component),
		Name: reflect.TypeOf(component).Name(),
	}

	var info, err = b.registry.Info(token)

	if err != nil {
		b.err = err
		return err
	}

	if _, exists := b.seen[info.Id]; exists {
		b.err = fmt.Errorf("%w: %s", ErrDuplicateComponent, info.Name)
		return b.err
	}

	b.seen[info.Id] = struct{}{}
	b.components = append(b.components, component)

	return nil
}

func (b *BundleBuilder) Components() ([]any, error) {
	return b.components, b.err
}

func SpawnBundle(world *World, bundle Bundle) (Entity, error) {
	if bundle == nil {
		return Entity{}, ErrInvalidComponentType
	}

	var builder = &BundleBuilder{
		components: make([]any, 0),
		seen:       make(map[ComponentID]struct{}),
		registry:   world.registry,
	}

	if err := bundle.Apply(builder); err != nil {
		return Entity{}, err
	}

	var components, err = builder.Components()
	if err != nil {
		return Entity{}, err
	}

	return Spawn(world, components...)
}
