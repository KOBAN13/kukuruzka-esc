package ecs

import (
	"errors"
	"reflect"
)

type CommandBuffer struct {
	commands []Command
	errors   []error
}

type Command interface {
	apply(world *World) error
}

func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		commands: make([]Command, 0),
		errors:   make([]error, 0),
	}
}

func (c *CommandBuffer) Spawn() *SpawnCommandBuilder {
	return &SpawnCommandBuilder{
		buffer:     c,
		components: make([]any, 0),
		seen:       make(map[reflect.Type]struct{}),
	}
}

func (c *CommandBuffer) Add(entity Entity) *AddCommandBuilder {
	return &AddCommandBuilder{
		buffer:     c,
		components: make([]any, 0),
		seen:       make(map[reflect.Type]struct{}),
		entity:     entity,
	}
}

func (c *CommandBuffer) Remove(entity Entity) *RemoveCommandBuilder {
	return &RemoveCommandBuilder{
		buffer:           c,
		entity:           entity,
		componentsTokens: make([]ComponentToken, 0),
		seen:             make(map[reflect.Type]struct{}),
	}
}

func (c *CommandBuffer) Despawn(entity Entity) error {
	c.add(despawnCommand{
		entity: entity,
	})

	return nil
}

func (c *CommandBuffer) Apply(world *World) error {
	if len(c.errors) > 0 {
		return errors.Join(c.errors...)
	}

	var prev = world.mutationPhase
	world.mutationPhase = MutationApplyingCommands

	defer func() {
		world.mutationPhase = prev
	}()

	for _, command := range c.commands {
		if err := command.apply(world); err != nil {
			return err
		}
	}

	return nil
}

func (c *CommandBuffer) Clear() {
	clear(c.commands)
	c.commands = c.commands[:0]
	clear(c.errors)
	c.errors = c.errors[:0]
}

func (c *CommandBuffer) Len() int {
	return len(c.commands)
}

func (c *CommandBuffer) Errors() []error {
	return c.errors
}

func (c *CommandBuffer) add(command Command) {
	c.commands = append(c.commands, command)
}
