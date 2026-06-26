package ecs

import (
	"errors"
	"fmt"
	"reflect"
)

type RemoveCommandBuilder struct {
	buffer           *CommandBuffer
	entity           Entity
	componentsTokens []ComponentToken
	err              error
	commited         bool
	seen             map[reflect.Type]struct{}
}

type removeCommand struct {
	entity     Entity
	components []ComponentToken
}

func (b *RemoveCommandBuilder) Component(component ComponentToken) *RemoveCommandBuilder {
	if b.err != nil {
		return b
	}

	if b.commited {
		b.err = errors.New("remove command already committed")
		return b
	}

	var typeComponent = component.Type

	if err := validateComponentType(typeComponent); err != nil {
		b.err = err
		return b
	}

	if _, exists := b.seen[typeComponent]; exists {
		b.err = fmt.Errorf("%w: %s", ErrDuplicateComponent, typeComponent.Name())
		return b
	}

	b.seen[typeComponent] = struct{}{}
	b.componentsTokens = append(b.componentsTokens, component)

	return b
}

func (b *RemoveCommandBuilder) Commit() error {
	if b.err != nil {
		return b.err
	}

	if b.commited {
		return errors.New("remove command already committed")
	}

	if len(b.componentsTokens) == 0 {
		return errors.New("remove command requires at least one component")
	}

	b.commited = true

	b.buffer.add(removeCommand{
		entity:     b.entity,
		components: append([]ComponentToken(nil), b.componentsTokens...),
	})

	return nil
}

func (c removeCommand) apply(world *World) error {
	for _, component := range c.components {
		var err = Remove(world, c.entity, component)

		if errors.Is(err, ErrComponentNotFound) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}
