package folder

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"go-micloud/internal/file"
	"go-micloud/pkg/color"
	"go-micloud/pkg/utils"
	"strings"
)

const (
	Tfile   = "file"
	Tfolder = "folder"
)

type Folder struct {
	Cursor *file.File
	Root   *file.File
}

func NewFolder() *Folder {
	base := &file.File{
		Name:     "/",
		Id:       "0",
		Type:     Tfolder,
		Revision: "",
		Parent:   nil,
		Child:    nil,
	}
	return &Folder{
		Cursor: base,
		Root:   base,
	}
}

// 打印树型目录结构
func (r *Folder) PrintFolder(root *file.File, level int) {
	if level > 0 {
		fmt.Println(strings.Repeat("│   ", level-1) + strings.Repeat("├── ", 1) + root.Name)
	} else {
		fmt.Println(strings.Repeat("├── ", level) + root.Name)
	}
	level++
	if root.Child != nil {
		for _, v := range root.Child {
			r.PrintFolder(v, level)
		}
	}
}

// 改变当前目录
func (r *Folder) ChangeFolder(name string) error {
	if name == "/" {
		r.Cursor = r.Root
	} else if strings.HasPrefix(name, "..") {
		count := strings.Count(name, "..")
		for i := 0; i < count; i++ {
			if r.Cursor.Parent == nil {
				break
			}
			r.Cursor = r.Cursor.Parent
		}
	} else {
		if r.Cursor.Child != nil {
			for _, v := range r.Cursor.Child {
				if name == v.Name && v.Type == Tfolder {
					r.Cursor = v
				}
			}
		}
		if r.Cursor.Name != name {
			return errors.New("目录不存在")
		}
	}
	return nil
}

func (r *Folder) AddFolder(files []*file.File) {
	if r.Cursor.Name == "/" {
		r.Cursor = r.Root
	}
	for _, old := range r.Cursor.Child {
		if old.Child != nil {
			for _, f := range files {
				if f.Name == old.Name {
					f.Child = old.Child
				}
			}
		}
	}
	for _, f := range files {
		f.Parent = r.Cursor
	}
	r.Cursor.Child = files
}

func (r *Folder) Format() {
	fmt.Printf("total %d\n", len(r.Cursor.Child))
	for _, v := range r.Cursor.Child {
		if v.Type == "file" {
			fmt.Printf("- | %6s | %s | %s\n", humanize.Bytes(uint64(v.Size)), utils.FormatTimeInt(int64(v.CreateTime), true), v.Name)
		} else {
			fmt.Printf("d | ------ | %s | %s\n", utils.FormatTimeInt(int64(v.CreateTime), true), color.Blue(v.Name))
		}
	}
}
