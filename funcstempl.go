package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image"
	_ "image/jpeg"
	"os"
	"strings"
)

// default function map for templates
var funcMap = template.FuncMap{
	"domain":        getDomain,
	"langAttribute": langAttribute,
	"linkEntry":     linkEntry,
	"workEntry":     workEntry,
	"header":        header,
	"footer":        footer,
	"getPortrait":   getPortrait,
	"getComment":    getComment,
	"fmtYear":       fmtYear,
}

func getDomain() template.URL {
	return template.URL(address)
}

func formatLangs(allLangs string) template.HTML {
	langArr := strings.Split(allLangs, ",")
	for i, l := range langArr {
		langArr[i] = fmt.Sprintf("<span class=\"langcode\">%s</span>", l)
	}
	return template.HTML(strings.Join(langArr, "\n"))
}

func langAttribute(langCode string) template.HTMLAttr {
	if langCode == "en" {
		return ""
	}
	if langCode == "grc" || langCode == "ru" {
		langCode += "-Latn"
	}
	return template.HTMLAttr(fmt.Sprintf(" lang=\"%s\"", langCode))
}

func linkEntry(data Link) (template.HTML, error) {
	linkTmpl := template.Must(template.New("linkEntry.tmpl").ParseFiles(path(templatesDir, partialsDir, "linkEntry.tmpl")))
	bfr := bytes.Buffer{}
	err := linkTmpl.Execute(&bfr, data)
	return template.HTML(bfr.String()), err
}

func workEntry(data Work) (template.HTML, error) {
	funcMap := template.FuncMap{"fmtYear": fmtYear, "formatLangs": formatLangs, "langAttribute": langAttribute, "domain": getDomain}
	workTmpl := template.Must(template.New("workEntry.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, partialsDir, "workEntry.tmpl")))
	bfr := bytes.Buffer{}
	err := workTmpl.Execute(&bfr, data)
	return template.HTML(bfr.String()), err
}

func header(title string) (template.HTML, error) {
	funcMap := template.FuncMap{"domain": getDomain}
	hdrTmpl := template.Must(template.New("header.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, partialsDir, "header.tmpl")))
	bfr := bytes.Buffer{}
	err := hdrTmpl.Execute(&bfr, title)
	return template.HTML(bfr.String()), err
}

func footer() (template.HTML, error) {
	ftr, err := os.ReadFile(path(templatesDir, partialsDir, "footer.html"))
	return template.HTML(ftr), err
}

func getPortrait(slug string) (template.HTML, error) {
	portrait, err := os.Open("static/portraits/" + slug + ".jpg")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	// get portrait dimension
	imgConfig, _, err := image.DecodeConfig(portrait)
	if err != nil {
		return "", err
	}
	funcMap := template.FuncMap{"domain": getDomain}
	ptrtTmpl := template.Must(template.New("portrait.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, partialsDir, "portrait.tmpl")))
	bfr := bytes.Buffer{}
	data := struct {
		Path   string
		Width  int
		Height int
	}{
		Path:   slug + ".jpg",
		Width:  imgConfig.Width,
		Height: imgConfig.Height,
	}
	err = ptrtTmpl.Execute(&bfr, data)
	return template.HTML(bfr.String()), err
}

func getComment(slug string) (template.HTML, error) {
	comment, err := os.ReadFile("comments/" + slug + ".html")
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return template.HTML(comment), err
}

func fmtYear(year int) template.HTML {
	if year < 0 {
		return template.HTML(fmt.Sprintf("%d BC", -year))
	}
	return template.HTML(fmt.Sprint(year))
}
