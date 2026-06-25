package ecs

type archetypeID uint32

type archetype struct {
	id        archetypeID
	signature componentSignature
	entities  []Entity
	columns   map[ComponentID]column
}

func newArchetype(id archetypeID, signature componentSignature, columns map[ComponentID]column) *archetype {
	return &archetype{
		id:        id,
		signature: signature,
		entities:  make([]Entity, 0),
		columns:   columns,
	}
}

func (a *archetype) Len() int {
	return len(a.entities)
}

func (a *archetype) appendEntity(entity Entity, values map[ComponentID]any) (int, error) {
	var row = a.Len()

	a.entities = append(a.entities, entity)
	var appended = make([]column, 0, len(a.columns))

	for _, componentId := range a.signature.ids {
		var col = a.columns[componentId]

		if value, ok := values[componentId]; ok {
			if err := col.AppendAny(value); err != nil {
				a.entities = a.entities[:row]

				for i := len(appended) - 1; i >= 0; i-- {
					appended[i].SwapRemove(row)
				}

				return 0, err
			}
		} else {
			col.AppendZero()
		}

		appended = append(appended, col)
	}

	return row, nil
}

func (a *archetype) removeEntity(row int) (moved Entity, movedRow int, hadMove bool) {
	var last = a.Len() - 1

	if row != last {
		moved = a.entities[last]
		a.entities[row] = moved
		movedRow = row
		hadMove = true
	}

	for _, col := range a.columns {
		col.SwapRemove(row)
	}

	a.entities = a.entities[:last]

	return moved, movedRow, hadMove
}

func (a *archetype) column(id ComponentID) (column, bool) {
	if col, ok := a.columns[id]; ok {
		return col, true
	}

	return nil, false
}
