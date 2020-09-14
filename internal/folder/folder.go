package folder

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"go-micloud/internal/file"
	"go-micloud/pkg/color"
	"go-micloud/pkg/line"
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
func PrintFolder(root *file.File, level int) {
	if level > 0 {
		fmt.Println(strings.Repeat("│   ", level-1) + strings.Repeat("├── ", 1) + root.Name)
	} else {
		fmt.Println(strings.Repeat("├── ", level) + root.Name)
	}
	level++
	if root.Child != nil {
		for _, v := range root.Child {
			PrintFolder(v, level)
		}
	}
}

func ChangeFolder(folder *Folder, name string) error {
	if name == "/" {
		folder.Cursor = folder.Root
	} else if strings.HasPrefix(name, "..") {
		count := strings.Count(name, "..")
		for i := 0; i < count; i++ {
			if folder.Cursor.Parent == nil {
				break
			}
			folder.Cursor = folder.Cursor.Parent
		}
	} else {
		if folder.Cursor.Child != nil {
			for _, v := range folder.Cursor.Child {
				if name == v.Name && v.Type == Tfolder {
					folder.Cursor = v
				}
			}
		}
		if folder.Cursor.Name != name {
			return errors.New("目录不存在")
		}
	}
	setUpLinePrefix(folder.Cursor)
	go setUpWordCompleter(folder.Cursor.Child)
	return nil
}

func AddFolder(folder *Folder, files []*file.File) {
	if folder.Cursor.Name == "/" {
		folder.Cursor = folder.Root
	}
	for _, old := range folder.Cursor.Child {
		if old.Child != nil {
			for _, f := range files {
				if f.Name == old.Name {
					f.Child = old.Child
				}
			}
		}
	}
	for _, f := range files {
		f.Parent = folder.Cursor
	}
	folder.Cursor.Child = files
	go setUpWordCompleter(files)
}

func Format(files []*file.File) {
	fmt.Printf("total %d\n", len(files))
	for _, v := range files {
		if v.Type == "file" {
			fmt.Printf("- | %-6s | %s | %s\n", humanize.Bytes(uint64(v.Size)), utils.FormatTimeInt(int64(v.CreateTime), true), v.Name)
		} else {
			fmt.Printf("d | ------ | %s | %s\n", utils.FormatTimeInt(int64(v.CreateTime), true), color.Blue(v.Name))
		}
	}
}

// 设置Tab补全提示
func setUpWordCompleter(files []*file.File) {
	var completerWord []string
	for _, f := range files {
		completerWord = append(completerWord, f.Name)
	}
	line.CsLiner.SetWorldCompleter(completerWord)
}

func setUpLinePrefix(cursor *file.File) {
	var (
		names []string
		c     = cursor
	)
	for c != nil {
		if c.Name != "/" {
			names = append(names, c.Name)
		}
		c = c.Parent
	}
	var path string
	for i := len(names); i > 0; i-- {
		path = path + "/" + names[i-1]
	}
	line.CsLiner.SetUpPrefix(strings.ReplaceAll(path, "//", "/"))
}
