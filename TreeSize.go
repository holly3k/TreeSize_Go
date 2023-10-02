// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/lxn/walk"
	"github.com/skratchdot/open-golang/open"

	. "github.com/lxn/walk/declarative"
)


func main() {
	fileSizeTreeRun()
}
type FileSizeMainWindow struct {
	*walk.MainWindow
}
var statusBar *walk.StatusBarItem

func fileSizeTreeRun() {
	mw := new(FileSizeMainWindow)

	var openAction, goToFolderAction *walk.Action
	var treeView *walk.TreeView
	treeModel := emptyTreeModel()
	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "File Size Tree",
		MenuItems: []MenuItem{
			Menu{
				Text: "&File",
				Items: []MenuItem{
					Action{
						AssignTo: &openAction,
						Text:     "&Open",
						Shortcut:    Shortcut{walk.ModControl, walk.KeyO},
						OnTriggered: func ()  {
							dlg := new(walk.FileDialog)

							if ok, err := dlg.ShowBrowseFolder(mw); err != nil {
								fmt.Println(err)
							} else if !ok {
								fmt.Println("not OK")
							} else {
								fmt.Println(dlg.FilePath)
								treeModel,_ = NewTreeModel(dlg.FilePath)
								treeView.SetModel(treeModel)
							}
						},
					},
					Separator{},
					Action{
						Text:        "E&xit",
						OnTriggered: func() { mw.Close() },
					},
				},
			},
		},
		MinSize: Size{300, 200},
		Layout:  VBox{},
		Children: []Widget{
			TreeView{
				AssignTo: &treeView,
				Model:    treeModel,
				ContextMenuItems: []MenuItem{
					Action{
						AssignTo:    &goToFolderAction,
						Text:        "Open",
						OnTriggered: func ()  {
							if value, ok := treeView.CurrentItem().(*FileAndSize); ok {
								open.Run(value.FullPath)
							  } else {
								fmt.Println("Type assertion failed.")
							  }
						},
					},
				},
			},
		},
		StatusBarItems: []StatusBarItem{
			StatusBarItem{
				AssignTo: &statusBar,
				Text:        "",
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	mw.Run()
}

type DirectoryTreeModel1 struct {
	walk.TreeModelBase
	roots []*FileAndSize
}

func emptyTreeModel() *DirectoryTreeModel1 {
	model := new(DirectoryTreeModel1)
	return model
}
func NewTreeModel(path string) (*DirectoryTreeModel1, error) {
	statusBar.SetText("start scanning....")
	model := new(DirectoryTreeModel1)
	
	model.roots = append(model.roots, scanDir(path,nil))
	statusBar.SetText("finish scanning")

	return model, nil
}

func (DirectoryTreeModel1) LazyPopulation() bool {
	// We don't want to eagerly populate our tree view with the whole file system.
	return true
}

func (m DirectoryTreeModel1) RootCount() int {
	return len(m.roots)
}

func (m DirectoryTreeModel1) RootAt(index int) walk.TreeItem {
	return m.roots[index]
}



type FileAndSize struct {
	FullPath     string         `json:"path"`
	Size         float64        `json:"size"`
	SizeReadable string         `json:"readableSize"`
	Childs       []*FileAndSize `json:"childs"`
	parent       *FileAndSize
	Type         string `json:"type"`
}

// ChildAt implements walk.TreeItem.
func (c FileAndSize) ChildAt(index int) walk.TreeItem {
	return c.Childs[index]
}

// ChildCount implements walk.TreeItem.
func (c FileAndSize) ChildCount() int {
	if c.Childs == nil {
		return 0
	}
	return len(c.Childs)
}

// Parent implements walk.TreeItem.
func (c FileAndSize) Parent() walk.TreeItem {
	if c.parent == nil {
		return nil
	}
	return c.parent
}

// Text implements walk.TreeItem.
func (c FileAndSize) Text() string {
	if(c.parent == nil) {
		return c.FullPath + "("+c.SizeReadable+")"
	}
	fileInfor, err := os.Stat(c.FullPath)
	if err != nil {
		fmt.Println(err)
	}
	return fileInfor.Name() + "("+c.SizeReadable+")"
}

func (d FileAndSize) Image() interface{} {
	return d.FullPath
}

func scanDir(path string, parent *FileAndSize) *FileAndSize {

	dirAry, err := ioutil.ReadDir(path)
	folderDetail := FileAndSize{}
	folderDetail.FullPath = path
	folderDetail.Size = 0
	folderDetail.Type = "FOLDER"
	folderDetail.parent = parent
	folderDetail.SizeReadable = "0kb"
	folderDetail.Childs = []*FileAndSize{}
	if err != nil {
		return &folderDetail
	}
	for _, e := range dirAry {
		if e.IsDir() {
			subFolder := scanDir(filepath.Join(path, e.Name()),&folderDetail)
			folderDetail.Childs = append(folderDetail.Childs, subFolder)
			folderDetail.Size += subFolder.Size
		} else {
			folderDetail.Size += float64(e.Size())
			folderDetail.Childs = append(folderDetail.Childs, &FileAndSize{parent:&folderDetail, Type: "FILE", Childs: []*FileAndSize{}, FullPath: filepath.Join(path, e.Name()), Size: float64(e.Size()), SizeReadable: formatSize(float64(e.Size()))})
			// go statusBar.SetText(filepath.Join(path, e.Name()))
		}
	}
	folderDetail.SizeReadable = formatSize(folderDetail.Size)
	sort.Slice(folderDetail.Childs, func(i, j int) bool {
		return folderDetail.Childs[i].Size > folderDetail.Childs[j].Size
	})
	return &folderDetail
}

func formatSize(size float64) string {
	if size < 1024 {
		return strconv.FormatFloat(size, 'f', 2, 32) + "bit"
	} else {
		size = size / 1024
		if size < 1024 {
			return strconv.FormatFloat(size, 'f', 2, 32) + "k"
		} else {
			size = size / 1024
			if size < 1024 {
				return strconv.FormatFloat(size, 'f', 2, 32) + "M"
			} else {
				size = size / 1024
				if size < 1024 {
					return strconv.FormatFloat(size, 'f', 2, 32) + "G"
				} else {
					size = size / 1024
					if size < 1024 {
						return strconv.FormatFloat(size, 'f', 2, 32) + "T"
					} else {
						size = size / 1024
						return strconv.FormatFloat(size, 'f', 2, 32) + "P"
					}
				}
			}
		}
	}
}