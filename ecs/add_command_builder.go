package ecs

import (
	"errors"
	"fmt"
	"reflect"
)

type AddCommandBuilder struct {
	buffer     *CommandBuffer
	entity     Entity
	components []any
	seen       map[reflect.Type]struct{}
	err        error
	commited   bool
}

type addCommand struct {
	entity     Entity
	components []any
}

func (b *AddCommandBuilder) With(component any) *AddCommandBuilder {
	if b.err != nil {
		return b
	}

	if b.commited {
		b.err = errors.New("add command already committed")
		return b
	}

	var typeComponent = reflect.TypeOf(component)

	if err := validateComponentType(typeComponent); err != nil {
		b.err = err
		return b
	}

	if _, exists := b.seen[typeComponent]; exists {
		b.err = fmt.Errorf("%w: %s", ErrDuplicateComponent, typeComponent.Name())
		return b
	}

	b.seen[typeComponent] = struct{}{}
	b.components = append(b.components, component)

	return b
}

func (b *AddCommandBuilder) Bundle(bundle Bundle) *AddCommandBuilder {
	if b.err != nil {
		return b
	}

	if b.commited {
		b.err = errors.New("add command already committed")
		return b
	}

	var builder = NewBundleBuilder()

	if err := bundle.Apply(builder); err != nil {
		b.err = err
		return b
	}

	var components, err = builder.Components()

	if err != nil {
		b.err = err
		return b
	}

	for _, component := range components {
		b.With(component)
	}

	return b
}

func (b *AddCommandBuilder) Commit() error {
	if b.err != nil {
		return b.err
	}

	if b.commited {
		return errors.New("add command already committed")
	}

	if len(b.components) == 0 {
		return errors.New("add command requires at least one component")
	}

	b.commited = true

	b.buffer.add(addCommand{
		entity:     b.entity,
		components: append([]any(nil), b.components...),
	})

	return nil
}

func (c addCommand) apply(world *World) error {
	return Add(world, c.entity, c.components...)
}
