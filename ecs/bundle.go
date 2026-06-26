package ecs

import (
	"fmt"
	"reflect"
)

type BundleBuilder struct {
	components []any
	seen       map[reflect.Type]struct{}
	err        error
}

type Bundle interface {
	Apply(*BundleBuilder) error
}

type BundleFunc func(*BundleBuilder) error

func NewBundleBuilder() *BundleBuilder {
	return &BundleBuilder{
		components: make([]any, 0),
		seen:       make(map[reflect.Type]struct{}),
	}
}

func (f BundleFunc) Apply(builder *BundleBuilder) error {
	return f(builder)
}

func (b *BundleBuilder) With(component any) error {
	var componentType = reflect.TypeOf(component)

	if err := validateComponentType(componentType); err != nil {
		b.err = err
		return err
	}

	if _, exists := b.seen[componentType]; exists {
		b.err = fmt.Errorf("%w: %s", ErrDuplicateComponent, componentType.Name())
		return b.err
	}

	b.seen[componentType] = struct{}{}
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

	var builder = NewBundleBuilder()

	if err := bundle.Apply(builder); err != nil {
		return Entity{}, err
	}

	var components, err = builder.Components()
	if err != nil {
		return Entity{}, err
	}

	return Spawn(world, components...)
}
