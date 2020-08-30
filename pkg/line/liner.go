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
		spaceIndex := strings.LastIndex(line, " ")
		if spaceIndex >= 0 {
			prefix = line[spaceIndex+1:]
			command = line[:spaceIndex]
		}
		var candidates []string

		if command == "upload" {
			var dir string
			if strings.HasPrefix(prefix, "/") {
				dir = path.Dir(prefix)
				dir = strings.ReplaceAll(dir, "\\s", " ")
				prefix = prefix[strings.LastIndex(prefix, "/")+1:]
			} else {
				dir, _ = os.Getwd()
			}
			infos, _ := ioutil.ReadDir(dir)
			for _, v := range infos {
				if strings.HasPrefix(strings.ToLower(v.Name()), strings.ToLower(prefix)) {
					name := strings.ReplaceAll(v.Name(), " ", "\\s")
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
					k = strings.ReplaceAll(k, " ", "\\s")
					candidates = append(candidates, " "+k)
				}
			}
		}
		return command, candidates, ""
	})
}
