package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func processCSVFile(filename string, noHeader bool, t *template.Template) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ','
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	var header []string

	switch len(records) {
	case 0:
		return
	case 1:
		if !noHeader {
			return
		}

		t.Execute(os.Stdout, records[0])
	default:
		if noHeader {
			for _, tuple := range records {
				t.Execute(os.Stdout, tuple)
			}
			return
		}

		header = records[0]
		for i, h := range header {
			header[i] = strings.Map(func(r rune) rune {
				switch r {
				case ' ', '\t', '\n', '\r':
					return -1
				case '/', '\\', '-', '+':
					return '_'
				default:
					return r
				}

			}, strings.ToLower(h))
		}

		tupleMap := make(map[string]any, len(header))
		data := records[1:]
		n := len(data)

		for i, tuple := range data {
			tupleMap["meta"] = map[string]any {
				"index": i,
				"first": i == 0,
				"last": i >= n-1, 
			}
			for i, h := range header {
				tupleMap[h] = tuple[i]
			}

			t.Execute(os.Stdout, tupleMap)
		}
	}
}

func main() {
	var templ string
	var noHeader bool

	flag.StringVar(&templ, "t", "", "output template")
	flag.BoolVar(&noHeader, "no-header", false, "input has no header line")
	flag.Parse()

	if templ == "" {
		_, _ = fmt.Fprint(os.Stderr, "no template")
		os.Exit(1)
	}

	t := template.Must(template.New("template").Funcs(map[string]any{
		"quote_literal": func (c string) string {
			return "'" + strings.ReplaceAll(c, "'", "''") + "'"
		},
		"newline": func () string {
			return "\n"
		},
		"nl": func () string {
			return "\n"
		},
	}).Parse(templ))

	for _, f := range flag.Args() {
		processCSVFile(f, noHeader, t)
	}
}
