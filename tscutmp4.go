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
	tvmodel *RowModel
	te      *walk.TextEdit
	tv      *walk.TableView
}

const tmp string = `tmp`

func main() {
	if _, err := os.Stat(tmp); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(tmp, 0666)
		}
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	mw := &MyMainWindow{tvmodel: NewRowModel()}

	ch := make(chan Row, 100)

	go func() {
		for item := range ch {

			copy(item.path, fmt.Sprintf("%s/%s", item.workdir, `input.ts`))
			exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\dgmpgdec158\DGIndex.exe`), `-hide`, `-IF=[input.ts]`, `-OM=2`, `-OF=[input.ts]`, `-AT=[G:\encode\encode_18_masako\template.avs]`, `-EXIT`})
			exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\BonTsDemux\BonTsDemuxC.exe`), `-i`, `input.ts`, `-o`, `input.ts.bontsdemux`, `-encode`, `Demux(wav)`, `-start`, `-quit`})
			exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\neroAacEnc.exe`), `-br`, `128000`, `-ignorelength`, `-if`, `input.ts.bontsdemux.wav`, `-of`, `input.ts.bontsdemux.aac`})
			exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\avs2wav.exe`), `input.ts.avs`, `input.ts.all.wav`})
			copy(filepath.Join(cwd, `extra\trim.avs`), item.workdir)
			copy(filepath.Join(cwd, `extra\aviutl.ini`), item.workdir)

			mw.tvmodel.items[item.index-1].status = Loaded
			mw.tvmodel.PublishRowChanged(item.index - 1)
		}
	}()

	if _, err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    fmt.Sprintf("tscutmp4"),
		MinSize:  Size{960, 480},
		Size:     Size{960, 480},
		Layout:   VBox{MarginsZero: true},
		OnDropFiles: func(files []string) {
			fmt.Println("-- dropped --")
			for _, f := range files {
				fmt.Println(f)
				file, err := os.Open(f)
				if err != nil {
					panic(err)
				}

				tvitem := Row{
					index:   mw.tvmodel.RowCount() + 1,
					path:    f,
					file:    file,
					workdir: fmt.Sprintf("%s/%03d", tmp, mw.tvmodel.RowCount()+1),
					status:  Loading,
				}

				err = os.Mkdir(abs(tvitem.workdir), 0666)
				if err != nil {
					panic(err)
				}
				mw.tvmodel.items = append(mw.tvmodel.items, tvitem)
				ch <- tvitem
			}
			mw.tvmodel.PublishRowsReset()

			fmt.Println("--")
		},
		Children: []Widget{
			VSplitter{
				Children: []Widget{
					TableView{
						AssignTo:   &mw.tv,
						CheckBoxes: true,
						Columns: []TableViewColumn{
							{Title: "index"},
							{Title: "path"},
							{Title: "workdir"},
							{Title: "status"},
						},
						Model: mw.tvmodel,
						OnCurrentIndexChanged: func() {
							i := mw.tv.CurrentIndex()
							if 0 <= i {
								fmt.Printf("OnCurrentIndexChange: %v\n", mw.tvmodel.items[i].path)
							}
						},
						OnItemActivated: mw.tv_ItemActivated,
					},
					TextEdit{
						AssignTo: &mw.te,
						ReadOnly: true,
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: false},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "EncodeChecked",
						OnClicked: func() {
							go func() {
								for _, item := range mw.tvmodel.items {
									if item.checked {
										fmt.Printf("encode start: %v\n", item)
										mw.tvmodel.items[item.index-1].status = Encoding
										mw.tvmodel.PublishRowChanged(item.index - 1)
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\avs2wav.exe`), `input.ts.avs`, `input.ts.wav`})
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\neroAacEnc.exe`), `-br`, `128000`, `-ignorelength`, `-if`, `input.ts.wav`, `-of`, `input.ts.aac`})
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\x264.32bit.0.130.22730.130.2273.exe`), `--threads`, `8`, `--scenecut`, `60`, `--crf`, `20`, `--level`, `3.1`, `--output`, `input.ts.mp4`, `input.ts.avs`})
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\mp4box.exe`), `-add`, `input.ts.mp4#video`, `-add`, `input.ts.aac#audio`, `-new`, `output.mp4`})
										os.Rename(filepath.Join(item.workdir, "output.mp4"), item.path+`.mp4`)
										fmt.Printf("encode end  : %v\n", item)
										mw.tvmodel.items[item.index-1].status = Encoded
										mw.tvmodel.PublishRowChanged(item.index - 1)
									}
								}
								fmt.Println()
							}()
						},
					},
					PushButton{
						Text: "Cancel",
					},
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
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

func (mw *MyMainWindow) tv_ItemActivated() {
	cwd, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	item := mw.tvmodel.items[mw.tv.CurrentIndex()]
	fmt.Println(item.path)
	go exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\aviutl99i8\aviutl.exe`), `trim.avs`, `-a`, `input.ts.all.wav`})
}
