package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type MyMainWindow struct {
	*walk.MainWindow
	model *EnvModel
	lb    *walk.ListBox
}

type EnvItem struct {
	name    string
	file    *os.File
	workdir string
	index   int
	status  int
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

	cwd, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	mw := &MyMainWindow{model: NewEnvModel()}

	ch := make(chan EnvItem, 100)

	go func() {
		for item := range ch {

			copy(item.name, fmt.Sprintf("%s/%s", item.workdir, `input.ts`))
			exec_cmd(item.workdir, []string{cwd + `\extra\dgmpgdec158\DGIndex.exe`, `-hide`, `-IF=[input.ts]`, `-OM=2`, `-OF=[input.ts]`, `-AT=[G:\encode\encode_18_masako\template.avs]`, `-EXIT`})
			exec_cmd(item.workdir, []string{cwd + `\extra\BonTsDemux\BonTsDemuxC.exe`, `-i`, `input.ts`, `-o`, `input.ts.bontsdemux`, `-encode`, `Demux(wav)`, `-start`, `-quit`})
			exec_cmd(item.workdir, []string{cwd + `\extra\neroAacEnc.exe`, `-br`, `128000`, `-ignorelength`, `-if`, `input.ts.bontsdemux.wav`, `-of`, `input.ts.bontsdemux.aac`})
			exec_cmd(item.workdir, []string{cwd + `\extra\avs2wav.exe`, `input.ts.avs`, `input.ts.all.wav`})
			copy(cwd+`\extra\trim.avs`, item.workdir)
			copy(cwd+`\extra\aviutl.ini`, item.workdir)

			mw.model.items[item.index-1].status = 1
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
				file, err := os.Open(f)
				if err != nil {
					panic(err)
				}

				item := EnvItem{
					name:    f,
					file:    file,
					workdir: fmt.Sprintf("%s/%03d", tmp, mw.model.ItemCount() + 1),
					index:   mw.model.ItemCount() + 1,
					status:  0,
				}

				err = os.Mkdir(abs(item.workdir), 0666)
				if err != nil {
					panic(err)
				}
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
							item := mw.model.items[mw.lb.CurrentIndex()]
							go exec_cmd(item.workdir, []string{cwd + `\extra\aviutl99i8\aviutl.exe`, `trim.avs`, `-a`, `input.ts.all.wav`})
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
	return fmt.Sprintf("%d : %03d : %s", m.items[index].status, m.items[index].index, m.items[index].file.Name())
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

func exec_cmd(dir string, cmd_and_args []string) error {
	cmd := exec.Command(cmd_and_args[0], cmd_and_args[1:]...)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func abs(path string) string {
	a, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return a
}
