package ecs

import "fmt"

type Iterator struct {
	query         *Query
	matchIndex    int
	row           int
	current       *archetype
	currentRow    int
	currentEntity Entity
	hasCurrent    bool
}

func (q *Query) Iter() *Iterator {
	if q.seenVersion != q.world.archetypeVersion {
		return q.refresh()
	}

	return q.newIterator()
}

func (q *Query) refresh() *Iterator {
	q.matches = q.matches[:0]

	for _, archetype := range q.world.archetypes {
		if !q.matchesArchetype(archetype) {
			continue
		}

		q.matches = append(q.matches, queryArchetypePlan{
			archetype: archetype,
		})
	}

	q.seenVersion = q.world.archetypeVersion

	return q.newIterator()
}

func (q *Query) newIterator() *Iterator {
	return &Iterator{
		query:      q,
		matchIndex: -1,
		row:        -1,
		currentRow: -1,
	}
}

func (q *Query) matchesArchetype(archetype *archetype) bool {
	if !archetype.signature.containsAll(q.descriptor.With) {
		return false
	}

	return archetype.signature.excludesAll(q.descriptor.Without)
}

func (it *Iterator) Next() bool {
	for {
		if it.current == nil || it.row+1 >= it.current.Len() {
			it.matchIndex++
			it.row = -1

			if it.matchIndex >= len(it.query.matches) {
				it.current = nil
				it.matchIndex = -1
				it.currentEntity = Entity{}
				it.hasCurrent = false
				return false
			}

			it.current = it.query.matches[it.matchIndex].archetype
			continue
		}

		it.row++
		it.currentRow = it.row
		it.currentEntity = it.current.entities[it.row]
		it.hasCurrent = true

		return true
	}
}

func (it *Iterator) Entity() Entity {
	if !it.hasCurrent {
		return Entity{}
	}

	return it.currentEntity
}

func Read[T any](it *Iterator) (T, error) {
	var zeroValue T
	var componentInfo, err = it.query.world.registry.Info(Component[T]())

	if err != nil {
		return zeroValue, err
	}

	if !containsComponent(componentInfo.Id, it.query.descriptor.Reads) {
		return zeroValue, fmt.Errorf("read: component %v does not contain read", componentInfo.Id)
	}

	var col, ok = it.current.column(componentInfo.Id)

	if !ok {
		return zeroValue, fmt.Errorf("read: component %v does not contain read", componentInfo.Id)
	}

	return col.ValueAny(it.row).(T), nil
}

func Write[T any](it *Iterator) (*T, error) {
	var zeroValue *T
	var componentInfo, err = it.query.world.registry.Info(Component[T]())

	if err != nil {
		return zeroValue, err
	}

	if !containsComponent(componentInfo.Id, it.query.descriptor.Writes) {
		return zeroValue, fmt.Errorf("read: component %v does not contain read", componentInfo.Id)
	}

	var col, ok = it.current.column(componentInfo.Id)

	if !ok {
		return zeroValue, fmt.Errorf("read: component %v does not contain read", componentInfo.Id)
	}

	return col.PtrAny(it.currentRow).(*T), nil
}
