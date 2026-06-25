package ecs

import (
	"fmt"
	"reflect"
	"strings"
)

type WorldOption func(*World)

type World struct {
	slots             []entitySlot
	freeEntityIndexes []uint32
	archetypes        []*archetype
	archetypeByKey    map[string]*archetype
	registry          *ComponentRegistry
	mutationPhase     MutationPhase
	archetypeVersion  uint64
}

func WithEntityCapacity(capacity int) WorldOption {
	return func(world *World) {
		world.slots = make([]entitySlot, 0, capacity)
	}
}

func NewWorld(options ...WorldOption) *World {
	var world = &World{
		archetypeByKey: make(map[string]*archetype),
		registry:       NewComponentRegistry(),
	}

	for _, option := range options {
		option(world)
	}

	return world
}

func Spawn(world *World, components ...any) (Entity, error) {
	if world.mutationPhase == MutationRunningSystem {
		return Entity{}, ErrInvalidMutationPhase
	}

	values, signature, err := world.collectComponentValues(components)
	if err != nil {
		return Entity{}, err
	}

	archetype, err := world.archetypeFor(signature)
	if err != nil {
		return Entity{}, err
	}

	entity := world.allocateEntity()

	row, err := archetype.appendEntity(entity, values)
	if err != nil {
		world.releaseEntity(entity)
		return Entity{}, err
	}

	world.slots[entity.index].location = entityLocation{
		archetype: archetype,
		row:       row,
	}

	return entity, nil
}

func Despawn(world *World, entity Entity) error {
	if world.mutationPhase == MutationRunningSystem {
		return ErrInvalidMutationPhase
	}

	var slot, err = world.validateAlive(entity)

	if err != nil {
		return err
	}

	world.removeFromCurrentArchetype(slot)
	world.releaseEntity(entity)

	return nil
}

func Add(world *World, entity Entity, components ...any) error {
	if world.mutationPhase == MutationRunningSystem {
		return ErrInvalidMutationPhase
	}

	slot, err := world.validateAlive(entity)

	if err != nil {
		return err
	}

	if len(components) == 0 {
		return nil
	}

	values, signature, err := world.collectComponentValues(components)

	if err != nil {
		return err
	}

	var source = slot.location.archetype

	if source == nil {
		return ErrInvalidEntity
	}

	for _, id := range signature.ids {
		if source.signature.containsComponent(id) {
			var info, _ = world.registry.InfoById(id)
			return fmt.Errorf("%w: %s", ErrDuplicateComponent, info.Name)
		}
	}

	var targetsId = make([]ComponentID, 0, len(source.signature.ids)+len(signature.ids))

	targetsId = append(targetsId, source.signature.ids...)
	targetsId = append(targetsId, signature.ids...)

	var targetSignature = newComponentSignature(targetsId)

	target, err := world.archetypeFor(*targetSignature)

	if err != nil {
		return err
	}

	return world.moveEntity(entity, target, values)
}

func Remove(world *World, entity Entity, componentTokens ...ComponentToken) error {
	if world.mutationPhase == MutationRunningSystem {
		return ErrInvalidMutationPhase
	}

	slot, err := world.validateAlive(entity)

	if err != nil {
		return err
	}

	if len(componentTokens) == 0 {
		return nil
	}

	var source = slot.location.archetype

	if source == nil {
		return ErrInvalidEntity
	}

	var removeIds = make([]ComponentID, 0, len(componentTokens))
	var seen = make(map[ComponentID]struct{}, len(componentTokens))

	for _, token := range componentTokens {
		info, err := world.registry.Info(token)

		if err != nil {
			return err
		}

		seen[info.Id] = struct{}{}

		if !source.signature.containsComponent(info.Id) {
			return fmt.Errorf("%w: %s", ErrComponentNotFound, info.Name)
		}

		removeIds = append(removeIds, info.Id)
	}

	var resultIds = make([]ComponentID, 0, len(source.signature.ids)-len(removeIds))

	for _, id := range source.signature.ids {
		if _, shouldDelete := seen[id]; !shouldDelete {
			resultIds = append(resultIds, id)
		}
	}

	var targetSignature = newComponentSignature(resultIds)
	target, err := world.archetypeFor(*targetSignature)

	if err != nil {
		return err
	}

	return world.moveEntity(entity, target, map[ComponentID]any{})
}

func Has[T any](world *World, entity Entity) (bool, error) {
	if world.mutationPhase == MutationRunningSystem {
		return false, ErrInvalidMutationPhase
	}

	slot, err := world.validateAlive(entity)

	if err != nil {
		return false, err
	}

	var source = slot.location.archetype

	if source == nil {
		return false, ErrInvalidEntity
	}

	info, err := world.registry.Info(Component[T]())

	if err != nil {
		return false, err
	}

	return source.signature.containsComponent(info.Id), nil
}

func Get[T any](world *World, entity Entity) (T, bool, error) {
	var zeroValue T

	if world.mutationPhase == MutationRunningSystem {
		return zeroValue, false, ErrInvalidMutationPhase
	}

	slot, err := world.validateAlive(entity)

	if err != nil {
		return zeroValue, false, err
	}

	var source = slot.location.archetype

	if source == nil {
		return zeroValue, false, ErrInvalidEntity
	}

	info, err := world.registry.Info(Component[T]())

	if err != nil {
		return zeroValue, false, err
	}

	col, ok := source.column(info.Id)

	if !ok {
		return zeroValue, false, nil
	}

	value, ok := col.ValueAny(slot.location.row).(T)

	if !ok {
		return zeroValue, false, fmt.Errorf(
			"%w: column get got %T, want %v",
			ErrInvalidComponentType,
			col.ValueAny(slot.location.row),
			info.Type,
		)
	}

	return value, true, nil
}

func GetWrite[T any](world *World, entity Entity) (*T, bool, error) {
	var zeroValue *T

	if world.mutationPhase == MutationRunningSystem {
		return zeroValue, false, ErrInvalidMutationPhase
	}

	slot, err := world.validateAlive(entity)

	if err != nil {
		return zeroValue, false, err
	}

	var source = slot.location.archetype
	if source == nil {
		return zeroValue, false, ErrInvalidEntity
	}

	info, err := world.registry.Info(Component[T]())

	if err != nil {
		return zeroValue, false, err
	}

	col, ok := source.column(info.Id)
	if !ok {
		return zeroValue, false, nil
	}

	value, ok := col.PtrAny(slot.location.row).(T)

	if !ok {
		return zeroValue, false, fmt.Errorf(
			"%w: column get got %T, want %v",
			ErrInvalidComponentType,
			col.PtrAny(slot.location.row),
			info.Type,
		)
	}

	return &value, true, nil
}

func Set[T any](world *World, entity Entity, value T) error {
	if world.mutationPhase == MutationRunningSystem {
		return ErrInvalidMutationPhase
	}

	slot, err := world.validateAlive(entity)

	if err != nil {
		return err
	}

	var source = slot.location.archetype

	if source == nil {
		return ErrInvalidEntity
	}

	info, err := world.registry.Info(Component[T]())

	if err != nil {
		return err
	}

	col, ok := source.column(info.Id)
	if !ok {
		return ErrInvalidComponentType
	}

	err = col.SetAny(slot.location.row, value)

	if err != nil {
		return err
	}

	return nil
}

func IsAlive(world *World, entity Entity) bool {
	if world.mutationPhase == MutationRunningSystem {
		return false
	}

	var slot, err = world.validateAlive(entity)

	if err != nil {
		return false
	}

	return slot.alive
}

func (world *World) InspectArchetypes() ArchetypeReport {
	var result ArchetypeReport

	for _, archetype := range world.archetypes {
		var archetypeInfo = ArchetypeInfo{
			ID:         uint32(archetype.id),
			Components: make([]string, len(archetype.entities)),
			Entities:   archetype.Len(),
		}

		for _, componentId := range archetype.signature.ids {
			var info, ok = world.registry.InfoById(componentId)

			if !ok {
				archetypeInfo.Components = append(
					archetypeInfo.Components,
					fmt.Sprintf("unknown archetype: %d", componentId))

				continue
			}

			archetypeInfo.Components = append(archetypeInfo.Components, info.Name)
		}
	}

	return result
}

func (world *World) DebugArchetypes() string {
	var archetypeReport = world.InspectArchetypes()

	var builder strings.Builder

	_, _ = fmt.Fprintf(&builder, "Archetypes: %d\n", len(archetypeReport.Archetypes))

	for _, archetype := range archetypeReport.Archetypes {
		var components = "<empty>"

		if len(archetype.Components) > 0 {
			components = strings.Join(archetype.Components, ", ")
		}

		_, _ = fmt.Fprintf(&builder, "\n#%d %s\n", archetype.ID, components)
		_, _ = fmt.Fprintf(&builder, "  entities: %d\n", archetype.Entities)
	}

	return builder.String()
}

func (world *World) allocateEntity() Entity {
	if n := len(world.freeEntityIndexes); n > 0 {
		var index = world.freeEntityIndexes[n-1]
		world.freeEntityIndexes = world.freeEntityIndexes[:n-1]
		var slot = &world.slots[index]
		slot.alive = true
		return Entity{index: index, generation: slot.generation}
	}

	var index = uint32(len(world.slots))
	world.slots = append(world.slots, entitySlot{alive: true})
	return Entity{index: index}
}

func (world *World) releaseEntity(entity Entity) {
	var slot = &world.slots[entity.index]
	slot.alive = false
	slot.location = entityLocation{}
	slot.generation++
	world.freeEntityIndexes = append(world.freeEntityIndexes, entity.index)
}

func (world *World) validateAlive(entity Entity) (*entitySlot, error) {
	if entity.IsZero() || int(entity.index) >= len(world.slots) {
		return nil, ErrInvalidEntity
	}

	slot := &world.slots[entity.index]

	if !slot.alive || slot.generation != entity.generation {
		return nil, ErrInvalidEntity
	}

	return slot, nil
}

func (world *World) collectComponentValues(components []any) (map[ComponentID]any, componentSignature, error) {
	var values = make(map[ComponentID]any, len(components))
	var ids = make([]ComponentID, 0, len(components))

	for _, component := range components {
		if component == nil {
			return nil, componentSignature{}, ErrInvalidComponentType
		}

		var token = ComponentToken{
			Type: reflect.TypeOf(component),
			Name: reflect.TypeOf(component).Name(),
		}

		var info, err = world.registry.Info(token)

		if err != nil {
			return nil, componentSignature{}, err
		}

		if _, exists := values[info.Id]; exists {
			return nil, componentSignature{}, fmt.Errorf("%w: %s", ErrDuplicateComponent, info.Name)
		}

		values[info.Id] = component
		ids = append(ids, info.Id)
	}

	var signature = newComponentSignature(ids)

	return values, *signature, nil
}

func (world *World) archetypeFor(signature componentSignature) (*archetype, error) {
	if archetype, ok := world.archetypeByKey[signature.key]; ok {
		return archetype, nil
	}

	var columns = make(map[ComponentID]column, len(signature.ids))

	for _, id := range signature.ids {
		var info, ok = world.registry.InfoById(id)

		if !ok {
			return nil, fmt.Errorf("%w: component id %d", ErrComponentNotFound, id)
		}

		columns[id] = newReflectColumn(info.Type)
	}

	archetype := newArchetype(
		archetypeID(len(world.archetypes)),
		signature,
		columns,
	)

	world.archetypes = append(world.archetypes, archetype)
	world.archetypeByKey[signature.key] = archetype
	world.archetypeVersion++

	return archetype, nil
}

func (world *World) moveEntity(entity Entity, target *archetype, values map[ComponentID]any) error {
	slot, err := world.validateAlive(entity)

	if err != nil {
		return err
	}

	var source = slot.location.archetype
	var sourceRow = slot.location.row

	var targetValues = make(map[ComponentID]any, len(target.signature.ids))

	for _, id := range target.signature.ids {
		if value, ok := values[id]; ok {
			targetValues[id] = value
			continue
		}

		var sourceColumn, ok = source.column(id)

		if !ok {
			continue
		}

		targetValues[id] = sourceColumn.ValueAny(sourceRow)
	}

	targetRow, err := target.appendEntity(entity, targetValues)

	if err != nil {
		return err
	}

	if source != nil {
		moved, movedRow, hadMove := source.removeEntity(sourceRow)
		if hadMove {
			world.slots[moved.index].location = entityLocation{
				archetype: source,
				row:       movedRow,
			}
		}
	}

	slot.location = entityLocation{
		archetype: target,
		row:       targetRow,
	}

	return nil
}

func (world *World) removeFromCurrentArchetype(slot *entitySlot) {
	var source = slot.location.archetype

	var moved, movedRow, hadMove = source.removeEntity(slot.location.row)

	if hadMove {
		world.slots[moved.index].location = entityLocation{
			archetype: source,
			row:       movedRow,
		}
	}

	slot.location = entityLocation{}
}
