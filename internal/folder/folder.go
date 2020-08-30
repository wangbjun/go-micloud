package folder

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"go-micloud/internal/api"
	"go-micloud/pkg/color"
	"go-micloud/pkg/function"
	"go-micloud/pkg/line"
	"strings"
)

const (
	Tfile   = "file"
	Tfolder = "folder"
)

type Folder struct {
	Cursor *api.File
	Root   *api.File
}

func NewFolder() *Folder {
	base := &api.File{
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

func PrintFolder(root *api.File, level int) {
	fmt.Println(strings.Repeat("  ", level) + root.Name)
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

func AddFolder(folder *Folder, name string, files []*api.File) {
	if name == "/" {
		folder.Root.Child = files
		for _, f := range files {
			f.Parent = folder.Root
		}
	} else {
		folder.Cursor.Child = files
		for _, f := range files {
			f.Parent = folder.Cursor
		}
	}
	go setUpWordCompleter(files)
}

func Format(files []*api.File) {
	fmt.Printf("total %d\n", len(files))
	for _, v := range files {
		if v.Type == "file" {
			fmt.Printf("- | %-6s | %s | %s\n", humanize.Bytes(uint64(v.Size)), function.FormatTimeInt(int64(v.CreateTime), true), v.Name)
		} else {
			fmt.Printf("d | ------ | %s | %s\n", function.FormatTimeInt(int64(v.CreateTime), true), color.Blue(v.Name))
		}
	}
}

// 设置Tab补全提示
func setUpWordCompleter(files []*api.File) {
	var completerWord []string
	for _, f := range files {
		completerWord = append(completerWord, strings.ReplaceAll(f.Name, " ", "\\s"))
	}
	line.CsLiner.SetWorldCompleter(completerWord)
}

func setUpLinePrefix(cursor *api.File) {
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
