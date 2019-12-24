package line

import (
	"github.com/peterh/liner"
	"io/ioutil"
	"path"
	"strings"
)

const base = "Go@MiCloud:~"

var CsLiner = Liner{liner.NewLiner(), base, 0}

type Liner struct {
	state  *liner.State
	prefix string
	n      int
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

func (l *Liner) AppendDir(dir string) {
	l.prefix = l.prefix + "/" + dir
	l.n++
}

func (l *Liner) RemoveDir(n int) {
	if n >= l.n || n == -1 {
		l.prefix = base
	} else {
		for i := 0; i < n; i++ {
			l.prefix = l.prefix[:strings.LastIndex(l.prefix, "/")]
		}
	}
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
		if strings.HasPrefix(prefix, "/") {
			dir := path.Dir(prefix)
			prefix = prefix[strings.LastIndex(prefix, "/")+1:]
			infos, _ := ioutil.ReadDir(strings.ReplaceAll(dir, "\\s", " "))
			for _, v := range infos {
				if strings.HasPrefix(v.Name(), prefix) {
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
				if strings.HasPrefix(k, prefix) {
					k = strings.ReplaceAll(k, " ", "\\s")
					candidates = append(candidates, " "+k)
				}
			}
		}
		return command, candidates, ""
	})
}
