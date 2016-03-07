package main

import (
	"github.com/lxn/walk"
	"os"
)

type Status int

const (
	Loading Status = iota
	Loaded
	Encoding
	Encoded
)

type Row struct {
	index   int
	path    string
	file    *os.File
	workdir string
	status  Status
	checked bool
}

type RowModel struct {
	walk.TableModelBase
	items []Row
}

func (m *RowModel) RowCount() int {
	return len(m.items)
}

func (m *RowModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.index
	case 1:
		return item.path
	case 2:
		return item.workdir
	case 3:
		return item.status
	}
	panic("unexpected col")
}

func (m *RowModel) Checked(row int) bool {
	return m.items[row].checked
}

func (m *RowModel) SetChecked(row int, checked bool) error {
	m.items[row].checked = checked
	return nil
}

func NewRowModel() *RowModel {
	m := new(RowModel)
	return m
}
