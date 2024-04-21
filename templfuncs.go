package main

import (
	"bytes"
	"errors"
	"html/template"
	"image"
	_ "image/jpeg"
	"os"
)

// default function map for templates
var funcMap = template.FuncMap{"domain": getDomain, "index": switchIndex, "header": header, "footer": footer, "getPortrait": getPortrait}

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
	ptrtTmpl := template.Must(template.New("portrait.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, "portrait.tmpl")))
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
