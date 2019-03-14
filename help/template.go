package help

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/jackdoe/baxx/common"
)

type HelpObject struct {
	Template      TextTemplate
	Err           error
	Status        *common.UserStatusOutput
	Notifications []common.PerRuleGroup
	FilePath      string
	Protected     bool
	Method        string
	SeeAlso       []HelpSeeAlso
	Token         string
	Email         string
}

type HelpSeeAlso struct {
	Method string
	Path   string
}

const (
	empty = ""
	tab   = "\t"
)

func PrettyJson(data interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(data)
	if err != nil {
		return empty, err
	}
	return buffer.String(), nil
}

func Load() *template.Template {
	fm := template.FuncMap{
		"add": func(a, b uint64) uint64 {
			return a + b
		},
		"prettySize": func(size uint64) string {
			return common.PrettySize(size)
		},
		"prettyFloat": func(p float64) string {
			return fmt.Sprintf("%.2f", p)
		},
		"pretty": func(data interface{}) string {
			p, err := PrettyJson(data)
			if err != nil {
				return err.Error()
			}
			return p
		},
	}
	r := os.Getenv("GOPATH")
	if r == "" {
		r = build.Default.GOPATH
	}
	t, err := template.New("root").Funcs(fm).ParseGlob(path.Join(r, "src", "github.com", "jackdoe", "baxx", "help", "t", "*.txt"))
	if err != nil {
		log.Fatal(err)
	}
	return t
}

//var root = Load()
func Render(data HelpObject) string {
	var out bytes.Buffer
	root := Load()
	err := root.ExecuteTemplate(&out, data.Template.String()+".txt", data)
	if err != nil {
		panic(err)
	}
	return out.String()
}
