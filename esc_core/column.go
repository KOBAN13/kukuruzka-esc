package esc_core

import (
	"fmt"
	"reflect"
)

type column interface {
	Len() int
	AppendZero()
	AppendAny(value any) error
	SwapRemove(row int)
	CopyValueTo(row int, targetColumn column) error
	ValueAny(row int) any
	PtrAny(row int) any
	SetAny(row int, value any) error
}

type reflectColumn struct {
	typ    reflect.Type
	values reflect.Value
}

func newReflectColumn(typ reflect.Type) *reflectColumn {
	return &reflectColumn{
		typ:    typ,
		values: reflect.MakeSlice(reflect.SliceOf(typ), 0, 0),
	}
}

func (c *reflectColumn) Len() int {
	return c.values.Len()
}

func (c *reflectColumn) AppendZero() {
	c.values = reflect.Append(c.values, reflect.Zero(c.typ))
}

func (c *reflectColumn) AppendAny(value any) error {
	got := reflect.TypeOf(value)
	if got != c.typ {
		return fmt.Errorf("%w: column append got %v, want %v", ErrInvalidComponentType, got, c.typ)
	}

	c.values = reflect.Append(c.values, reflect.ValueOf(value))
	return nil
}

func (c *reflectColumn) SwapRemove(row int) {
	last := c.values.Len() - 1

	c.values.Index(row).Set(c.values.Index(last))
	c.values.Index(last).Set(reflect.Zero(c.typ))
	c.values = c.values.Slice(0, last)
}

func (c *reflectColumn) CopyValueTo(row int, targetColumn column) error {
	return targetColumn.AppendAny(c.ValueAny(row))
}

func (c *reflectColumn) ValueAny(row int) any {
	return c.values.Index(row).Interface()
}

func (c *reflectColumn) PtrAny(row int) any {
	return c.values.Index(row).Addr().Interface()
}

func (c *reflectColumn) SetAny(row int, value any) error {
	got := reflect.TypeOf(value)
	if got != c.typ {
		return fmt.Errorf("%w: column set got %v, want %v", ErrInvalidComponentType, got, c.typ)
	}

	c.values.Index(row).Set(reflect.ValueOf(value))
	return nil
}
