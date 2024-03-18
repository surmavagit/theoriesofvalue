package main

import (
	"html/template"
	"os"
	"strings"
)

func uniquePage(tmplName string, funcMap template.FuncMap, data any, path string) error {
	indexTmpl := template.Must(template.New(tmplName).Funcs(funcMap).ParseFiles(templatesDir + "/" + tmplName))
	idx, err := os.Create(path)
	if err != nil {
		return err
	}
	return indexTmpl.Execute(idx, data)
}

func pageCollection(tmplName string, funcMap template.FuncMap, data map[string]any, commonPath string) error {
	authorTmpl := template.Must(template.New(tmplName).Funcs(funcMap).ParseFiles(templatesDir + "/" + tmplName))
	for slug, d := range data {
		err := os.Mkdir(strings.Join([]string{outputDir, commonPath, slug}, "/"), 0755)
		if err != nil {
			return err
		}
		f, err := os.Create(strings.Join([]string{outputDir, commonPath, slug, "index.html"}, "/"))
		if err != nil {
			return err
		}
		err = authorTmpl.Execute(f, d)
		if err != nil {
			return err
		}
	}
	return nil
}
