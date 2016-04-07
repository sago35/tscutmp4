package main

import (
	"bufio"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type MyMainWindow struct {
	*walk.MainWindow
	tvmodel *RowModel
	te      *walk.TextEdit
	tv      *walk.TableView
}

const tmp string = `tmp`

func main() {
	if i, err := os.Stat(tmp); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(tmp, 0666)
		}
	} else if i.IsDir() {
		if dirs, err := filepath.Glob(filepath.Join(tmp, "*")); err == nil {
			for _, dir := range dirs {
				if err := os.RemoveAll(dir); err != nil {
					panic(err)
				}
			}
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

			mklink_or_copy(item.path, fmt.Sprintf("%s/%s", item.workdir, `input.ts`))
			exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\dgmpgdec158\DGIndex.exe`), `-hide`, `-IF=[input.ts]`, `-OM=2`, `-OF=[input.ts]`, `-AT=[` + filepath.Join(cwd, `extra\template.avs`) + `]`, `-EXIT`})
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
					walk.MsgBox(mw.MainWindow, "Error", "フォルダを作成できません\r"+err.Error(), walk.MsgBoxIconError)
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
								mw.te.SetText(mw.tvmodel.items[i].trim)
							}
						},
						OnItemActivated: mw.tv_ItemActivated,
					},
					TextEdit{
						AssignTo: &mw.te,
						ReadOnly: false,
						OnKeyUp: func(key walk.Key) {
							i := mw.tv.CurrentIndex()
							if 0 <= i {
								mw.tvmodel.items[i].trim = mw.te.Text()
							}
						},
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
										trim := mw.tvmodel.items[item.index-1].trim
										trim = regexp.MustCompile(`[\r\n]+`).ReplaceAllString(trim, `++`)
										trim = regexp.MustCompile(`^\+\+`).ReplaceAllString(trim, ``)
										trim = regexp.MustCompile(`\+\+$`).ReplaceAllString(trim, ``)
										update_avs_file(filepath.Join(item.workdir, `input.ts.avs`), filepath.Join(item.workdir, `input.ts.after.avs`), trim)
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\avs2wav.exe`), `input.ts.after.avs`, `input.ts.wav`})
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\neroAacEnc.exe`), `-br`, `128000`, `-ignorelength`, `-if`, `input.ts.wav`, `-of`, `input.ts.aac`})
										exec_cmd(item.workdir, []string{filepath.Join(cwd, `extra\x264.32bit.0.130.22730.130.2273.exe`), `--threads`, `8`, `--scenecut`, `60`, `--crf`, `20`, `--level`, `3.1`, `--output`, `input.ts.mp4`, `input.ts.after.avs`})
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

func mklink_or_copy(src, dst string) {
	dst = strings.Replace(dst, `/`, `\`, -1)
	err := exec_cmd(`.`, []string{`cmd`, `/c`, `mklink`, `/h`, dst, src})
	if err != nil {
		fmt.Println("mklink error", err)
		copy(src, dst)
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

func update_avs_file(from, to, trim string) error {
	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	return update_avs(r, w, trim)
}

func update_avs(r io.Reader, w io.Writer, trim string) error {

	scanner := bufio.NewScanner(r)
	state := `NOT_SKIP`
	for scanner.Scan() {
		line := scanner.Text()
		if line == `# Trim() start` {
			state = `SKIP`
			fmt.Fprintf(w, "%s\r\n", scanner.Text())
		} else if line == `# Trim() end` {
			state = `NOT_SKIP`
			fmt.Fprintf(w, "%s\r\n", trim)
			fmt.Fprintf(w, "%s\r\n", scanner.Text())
		} else if state == `SKIP` {
		} else {
			fmt.Fprintf(w, "%s\r\n", scanner.Text())
		}
	}

	return nil
}
