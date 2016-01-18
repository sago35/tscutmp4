package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io"
	"log"
	"os"
	"path/filepath"
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

const tmp string = `tmp`

func main() {
	if _, err := os.Stat(tmp); err != nil {
		os.Mkdir(tmp, 0666)
	}

	mw := &MyMainWindow{model: NewEnvModel()}

	ch := make(chan EnvItem, 100)

	go func() {
		for item := range ch {
			dst := fmt.Sprintf("%s/%03d", tmp, item.index)
			os.Mkdir(dst, 0666)
			copy(item.name, dst)
			mw.model.items[item.index - 1].status = 1
			mw.lb.SetModel(mw.model)
		}
	}()

	if _, err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    fmt.Sprintf("tscutmp4"),
		MinSize:  Size{640, 240},
		Size:     Size{640, 240},
		Layout:   VBox{MarginsZero: true},
		OnDropFiles: func(files []string) {
			fmt.Println("-- dropped --")
			for _, f := range files {
				fmt.Println(f)
				item := EnvItem{name: f, index: mw.model.ItemCount() + 1, status: 0}
				mw.model.items = append(mw.model.items, item)
				ch <- item
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

func copy(src, dst string) {
	s, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	f, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			// 何もしない
		} else {
			panic(err)
		}
	} else {
		if f.IsDir() {
			dst = fmt.Sprintf("%s/%s", dst, filepath.Base(src))
		}
	}

	d, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		log.Fatal(err)
	}
}
