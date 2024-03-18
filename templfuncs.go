package main

import (
	"bytes"
	"html/template"
	"os"
)

// default function map for templates
var funcMap = template.FuncMap{"domain": getDomain, "index": switchIndex, "header": header, "footer": footer}

func relativeUrls() bool {
	args := os.Args
	return len(args) == 2 && args[1] == "-d"
}

func getDomain() template.URL {
	if relativeUrls() {
		lp, ok := os.LookupEnv("LOCALPATH")
		if !ok {
			panic("no localpath defined in .env file")
		}
		return template.URL(lp)
	}
	return template.URL("https://theoriesofvalue.com")
}

func switchIndex() template.URL {
	if relativeUrls() {
		return template.URL("/index.html")
	}
	return template.URL("/")
}

func header(title string) (template.HTML, error) {
	funcMap := template.FuncMap{"domain": getDomain, "index": switchIndex}
	hdrTmpl := template.Must(template.New("header.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, "header.tmpl")))
	bfr := bytes.Buffer{}
	err := hdrTmpl.Execute(&bfr, title)
	return template.HTML(bfr.String()), err
}

func footer() (template.HTML, error) {
	ftr, err := os.ReadFile(path(templatesDir, "footer.html"))
	return template.HTML(ftr), err
}
