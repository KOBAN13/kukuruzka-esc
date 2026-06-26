package ecs

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrResourcesNotFound    = errors.New("resources not found")
	ErrResourcesDuplicate   = errors.New("resources duplicate")
	ErrInvalidEntity        = errors.New("invalid entity")
	ErrComponentNotFound    = errors.New("component not found")
	ErrDuplicateComponent   = errors.New("duplicate component")
	ErrInvalidComponentType = errors.New("invalid component type")
	ErrInvalidMutationPhase = errors.New("invalid mutation phase")
	ErrAccessConflict       = errors.New("access conflict")
	ErrQueryAccess          = errors.New("query access violation")
	ErrInvalidComponentID   = errors.New("invalid component id")
)

type ComponentID uint32

type ComponentInfo struct {
	ID    ComponentID
	Name  string
	Type  reflect.Type
	Size  uintptr
	IsTag bool
}

type ComponentToken struct {
	Type reflect.Type
	Name string
}

type ComponentRegistry struct {
	next   ComponentID
	byType map[reflect.Type]ComponentInfo
	byId   map[ComponentID]ComponentInfo
}

func Component[T any]() ComponentToken {
	var componentType = reflect.TypeOf((*T)(nil)).Elem()

	return ComponentToken{
		Type: componentType,
		Name: componentType.Name(),
	}
}

func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		byType: make(map[reflect.Type]ComponentInfo),
		byId:   make(map[ComponentID]ComponentInfo),
	}
}

func (cr *ComponentRegistry) ID(token ComponentToken) (ComponentID, error) {
	var info, err = cr.Info(token)

	if err != nil {
		return 0, err
	}

	return info.ID, nil
}

func (cr *ComponentRegistry) Info(token ComponentToken) (ComponentInfo, error) {
	if err := validateComponentType(token.Type); err != nil {
		return ComponentInfo{}, err
	}

	if info, ok := cr.byType[token.Type]; ok {
		return info, nil
	}

	var id = cr.next
	cr.next++

	var info = ComponentInfo{
		ID:    id,
		Name:  token.Name,
		Type:  token.Type,
		Size:  token.Type.Size(),
		IsTag: token.Type.Size() == 0 && token.Type.NumField() == 0,
	}

	cr.byType[token.Type] = info
	cr.byId[id] = info
	return info, nil
}

func (cr *ComponentRegistry) InfoByID(id ComponentID) (ComponentInfo, bool) {
	var info, ok = cr.byId[id]

	return info, ok
}

func validateComponentType(componentType reflect.Type) error {
	if componentType == nil {
		return fmt.Errorf("%w: component type is nil", ErrInvalidComponentType)
	}

	if componentType.Kind() != reflect.Struct {
		return fmt.Errorf("%w: component type must be a struct", ErrInvalidComponentType)
	}

	return nil
}
