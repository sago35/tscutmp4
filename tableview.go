package main

import (
	//"fmt"
	"github.com/lxn/walk"
	//. "github.com/lxn/walk/declarative"
	//"sort"
)

type Row struct {
	Index int
	Path  string
}

type RowModel struct {
	walk.TableModelBase
	items []*Row
}

func (m *RowModel) RowCount() int {
	return len(m.items)
}

func (m *RowModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.Index
	case 1:
		return item.Path
	}
	panic("unexpected col")
}

func NewRowModel() *RowModel {
	m := new(RowModel)
	return m
}
