package main

import (
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

	err = os.Mkdir(outputDir+"/"+authorsDir, 0744)
	check(err)

	authorListTmpl := template.Must(template.ParseFiles(templatesDir + "/authorList.html"))
	al, err := os.Create(strings.Join([]string{outputDir, authorsDir, "index.html"}, "/"))
	check(err)
	err = authorListTmpl.Execute(al, data)
	check(err)

	authorTmpl := template.Must(template.ParseFiles(templatesDir + "/author.html"))
	for _, a := range data {
		err = os.Mkdir(strings.Join([]string{outputDir, authorsDir, a.Slug}, "/"), 0744)
		check(err)
		f, err := os.Create(strings.Join([]string{outputDir, authorsDir, a.Slug, "index.html"}, "/"))
		check(err)
		err = authorTmpl.Execute(f, a)
		check(err)
		fmt.Println(a.Slug)
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
