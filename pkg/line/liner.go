package line

import (
	"github.com/peterh/liner"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const base = "Go@MiCloud:"

var CsLiner = Liner{liner.NewLiner(), base}

var SysCommand = []string{"cd", "ls", "download", "upload", "login", "mkdir", "share", "tree", "rm"}

type Liner struct {
	state  *liner.State
	prefix string
}

func init() {
	CsLiner.state.SetCtrlCAborts(true)
}

func (l *Liner) Prompt() (string, error) {
	return l.state.Prompt(l.prefix + "$ ")
}

func (l *Liner) Close() error {
	return l.state.Close()
}

func (l *Liner) AppendHistory(item string) {
	l.state.AppendHistory(item)
}

func (l *Liner) SetUpPrefix(path string) {
	l.prefix = base + path
}

func (l *Liner) SetWorldCompleter(words []string) {
	l.state.SetWordCompleter(func(line string, pos int) (head string, completions []string, tail string) {
		var (
			prefix  = line
			command = line
		)
		firstIndex := strings.Index(line, " ")
		if firstIndex != -1 {
			command = line[:firstIndex]
			prefix = line[firstIndex+1:]
		}
		var candidates []string
		if command == "" {
			candidates = SysCommand
		} else if isSysCommand(command) {
			if command == "upload" {
				var dir string
				if strings.HasPrefix(prefix, "/") {
					dir = path.Dir(prefix)
					prefix = prefix[strings.LastIndex(prefix, "/")+1:]
				} else {
					dir, _ = os.Getwd()
				}
				infos, _ := ioutil.ReadDir(dir)
				for _, v := range infos {
					if strings.HasPrefix(v.Name(), ".") {
						continue
					}
					if strings.HasPrefix(strings.ToLower(v.Name()), strings.ToLower(prefix)) {
						name := v.Name()
						if v.IsDir() {
							name = name + "/"
						}
						candidate := " " + dir + "/" + name
						if dir == "/" {
							candidate = " /" + name
						}
						candidates = append(candidates, candidate)
					}
				}
			} else {
				for _, k := range words {
					if strings.HasPrefix(strings.ToLower(k), strings.ToLower(prefix)) {
						candidates = append(candidates, " "+k)
					}
				}
			}
		} else {
			for _, k := range SysCommand {
				if strings.HasPrefix(strings.ToLower(k), strings.ToLower(prefix)) {
					candidates = append(candidates, strings.ReplaceAll(k, prefix, "")+" ")
				}
			}
		}
		return command, candidates, ""
	})
}

func isSysCommand(cmd string) bool {
	for _, v := range SysCommand {
		if v == cmd {
			return true
		}
	}
	return false
}
