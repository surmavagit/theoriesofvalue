package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
)

var outputDir = "public"
var staticDir = "static"
var authorsDir = "authors"
var templatesDir = "templates"

func main() {
	// copy static files
	staticFiles, err := os.ReadDir(staticDir)
	check(err)
	for _, sf := range staticFiles {
		err := os.Link(staticDir+"/"+sf.Name(), outputDir+"/"+sf.Name())
		check(err)
	}

	funcMap := template.FuncMap{"domain": getDomain, "index": switchIndex, "header": header, "footer": footer}

	// generate main page
	indexTmpl := template.Must(template.New("index.tmpl").Funcs(funcMap).ParseFiles(templatesDir + "/index.tmpl"))
	idx, err := os.Create(strings.Join([]string{outputDir, "index.html"}, "/"))
	check(err)
	err = indexTmpl.Execute(idx, nil)
	check(err)

	// generate authors list
	authorListTmpl := template.Must(template.New("authorList.tmpl").Funcs(funcMap).ParseFiles(templatesDir + "/authorList.tmpl"))
	err = os.Mkdir(outputDir+"/"+authorsDir, 0755)
	check(err)
	al, err := os.Create(strings.Join([]string{outputDir, authorsDir, "index.html"}, "/"))
	check(err)
	err = authorListTmpl.Execute(al, data)
	check(err)

	// generate author pages
	authorTmpl := template.Must(template.New("author.tmpl").Funcs(funcMap).ParseFiles(templatesDir + "/author.tmpl"))
	for _, a := range data {
		err = os.Mkdir(strings.Join([]string{outputDir, authorsDir, a.Slug}, "/"), 0755)
		check(err)
		f, err := os.Create(strings.Join([]string{outputDir, authorsDir, a.Slug, "index.html"}, "/"))
		check(err)
		err = authorTmpl.Execute(f, a)
		check(err)
	}
}

func relativeUrls() bool {
	args := os.Args
	return len(args) == 2 && args[1] == "-d"
}

func getDomain() template.URL {
	if relativeUrls() {
		return template.URL(getEnv("LOCALPATH"))
	}
	return template.URL("https://theoriesofvalue.com")
}

func switchIndex() template.URL {
	if relativeUrls() {
		return template.URL("/index.html")
	}
	return template.URL("/")
}

func header(title string) template.HTML {
	funcMap := template.FuncMap{"domain": getDomain, "index": switchIndex}
	hdrTmpl := template.Must(template.New("header.tmpl").Funcs(funcMap).ParseFiles(templatesDir + "/header.tmpl"))
	bfr := bytes.Buffer{}
	err := hdrTmpl.Execute(&bfr, title)
	check(err)
	return template.HTML(bfr.String())
}

func footer() template.HTML {
	ftr, err := os.ReadFile(strings.Join([]string{templatesDir, "footer.html"}, "/"))
	check(err)
	return template.HTML(ftr)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
