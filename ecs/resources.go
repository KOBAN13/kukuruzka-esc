package ecs

import (
	"reflect"
)

type Resources struct {
	values map[reflect.Type]any
}

func NewResources() *Resources {
	return &Resources{
		values: make(map[reflect.Type]any),
	}
}

func PutResources[T any](resources *Resources, value T) {
	var key = reflect.TypeOf((*T)(nil)).Elem()

	resources.values[key] = value
}

func GetResources[T any](resources *Resources) (*T, bool, error) {
	var key = reflect.TypeOf((*T)(nil)).Elem()

	if _, ok := resources.values[key]; !ok {
		return nil, false, ErrComponentNotFound
	}

	var value, _ = resources.values[key].(*T)

	return value, true, nil
}

func RemoveResources[T any](resources *Resources) error {
	var key = reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := resources.values[key]; !ok {
		return ErrComponentNotFound
	}

	delete(resources.values, key)

	return nil
}
