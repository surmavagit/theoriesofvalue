package main

import (
	"html/template"
	"os"
	"strings"
)

type Page interface {
	getSlug() string
}

func (a Author) getSlug() string {
	return a.Slug
}
func (w Work) getSlug() string {
	return w.Slug
}

func uniquePage(tmplName string, funcMap template.FuncMap, data any, path string) error {
	indexTmpl := template.Must(template.New(tmplName).Funcs(funcMap).ParseFiles(templatesDir + "/" + tmplName))
	idx, err := os.Create(path)
	if err != nil {
		return err
	}
	return indexTmpl.Execute(idx, data)
}

func createPageInDir(p Page, tmpl *template.Template, commonPath string) error {
	slug := p.getSlug()
	err := os.Mkdir(strings.Join([]string{outputDir, commonPath, slug}, "/"), 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(strings.Join([]string{outputDir, commonPath, slug, "index.html"}, "/"))
	if err != nil {
		return err
	}
	return tmpl.Execute(f, p)
}
