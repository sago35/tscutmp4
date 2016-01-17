package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MyMainWindow struct {
	*walk.MainWindow
}

func main() {
	mw := &MyMainWindow{}

	MainWindow{
		AssignTo: &mw.MainWindow,
		Title:   fmt.Sprintf("hello walk"),
		MinSize: Size{320, 240},
		OnDropFiles: func(files []string) {
			fmt.Println("-- dropped --")
			for _, f := range files {
				fmt.Println(f)
			}
			fmt.Println("--")
		},
	}.Run()
}
