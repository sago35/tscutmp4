package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
)

type MyMainWindow struct {
	*walk.MainWindow
	model *EnvModel
	lb    *walk.ListBox
}

type EnvItem struct {
	name   string
	value  string
	index  int
	status int
}

type EnvModel struct {
	walk.ListModelBase
	items []EnvItem
}

func main() {
	mw := &MyMainWindow{model: NewEnvModel()}

	if _, err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    fmt.Sprintf("hello walk"),
		MinSize:  Size{320, 240},
		Size:     Size{320, 240},
		Layout:   VBox{MarginsZero: true},
		OnDropFiles: func(files []string) {
			fmt.Println("-- dropped --")
			for _, f := range files {
				fmt.Println(f)
				mw.model.items = append(mw.model.items, EnvItem{name: f, index: mw.model.ItemCount() + 1, status: 0})
			}
			mw.lb.SetModel(mw.model)

			fmt.Println("--")
		},
		Children: []Widget{
			VSplitter{
				Children: []Widget{
					ListBox{
						AssignTo: &mw.lb,
						Model:    mw.model,
						OnItemActivated: func() {
							walk.MsgBox(mw, "Info", mw.model.items[mw.lb.CurrentIndex()].name, walk.MsgBoxIconInformation)
						},
					},
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}

func NewEnvModel() *EnvModel {
	m := &EnvModel{items: make([]EnvItem, 0)}
	return m
}

func (m *EnvModel) ItemCount() int {
	return len(m.items)
}

func (m *EnvModel) Value(index int) interface{} {
	return fmt.Sprintf("%d : %03d : %s", m.items[index].status, m.items[index].index, m.items[index].name)
}
