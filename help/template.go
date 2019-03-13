package help

import (
	"bytes"
	"text/template"
)

func Parse(s string) *template.Template {
	fm := template.FuncMap{
		"add": func(a, b uint64) uint64 {
			return a + b
		},
	}
	t, err := template.New("main").Funcs(fm).Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}

func Render(t *template.Template, data interface{}) string {
	var out bytes.Buffer
	err := t.Execute(&out, data)

	if err != nil {
		panic(err)
	}
	return out.String()
}
