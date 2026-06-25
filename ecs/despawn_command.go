package ecs

import "errors"

type despawnCommand struct {
	entity Entity
}

func (c despawnCommand) apply(world *World) error {
	var err = Despawn(world, c.entity)

	if errors.Is(err, ErrInvalidEntity) {
		return nil
	}

	return err
}
