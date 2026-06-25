package ecs

import (
	"errors"
	"reflect"
)

type SpawnCommandBuilder struct {
	buffer     *CommandBuffer
	components []any
	seen       map[reflect.Type]struct{}
	err        error
	commited   bool
}

type spawnCommand struct {
	components []any
}

func (b *SpawnCommandBuilder) With(component any) *SpawnCommandBuilder {
	if b.err != nil {
		return b
	}

	if b.commited {
		b.err = errors.New("spawn command already committed")
		return b
	}

	var typeComponent = reflect.TypeOf(component)

	if err := validateComponentType(typeComponent); err != nil {
		b.err = err
		return b
	}

	if _, exists := b.seen[typeComponent]; exists {
		b.err = errors.New("spawn command already committed")
		return b
	}

	b.seen[typeComponent] = struct{}{}
	b.components = append(b.components, component)

	return b
}

func (b *SpawnCommandBuilder) Bundle(bundle Bundle) *SpawnCommandBuilder {
	if b.err != nil {
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

func (b *SpawnCommandBuilder) Commit() error {
	if b.err != nil {
		return b.err
	}

	if b.commited {
		return errors.New("spawn command already committed")
	}

	if len(b.components) == 0 {
		return errors.New("spawn command requires at least one component")
	}

	b.commited = true
	var components = append([]any(nil), b.components...)
	b.buffer.add(spawnCommand{components})

	return nil
}

func (c spawnCommand) apply(world *World) error {
	_, err := Spawn(world, c.components...)
	return err
}
