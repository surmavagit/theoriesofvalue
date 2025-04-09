package main

import (
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

const (
	outputDir    = "public"
	staticDir    = "static"
	templatesDir = "templates"
	partialsDir  = "partials"
	authorsDir   = "authors"
	worksDir     = "works"
	textsDir     = "read"
	portraitDir  = "portraits"
	siteLang     = "eng"
)

var (
	address = "https://theoriesofvalue.com"
	port    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	// load environment vars
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't load .env file: %s", err)
		return 1
	}

	// check args
	argNumber := len(os.Args)
	if argNumber > 3 {
		fmt.Fprint(os.Stderr, "incorrect number of arguments")
		return 1
	}

	if argNumber > 1 && os.Args[1] != "-p" {
		fmt.Fprint(os.Stderr, "incorrect option provided")
		return 1
	}

	// get default port from .env file, if needed
	if argNumber == 2 {
		lport, ok := os.LookupEnv("LOCALPORT")
		if !ok {
			fmt.Fprint(os.Stderr, "no LOCALPATH defined in .env file")
			return 1
		}
		port = lport
	}

	// get port number from provided argument, if needed
	if argNumber == 3 {
		onlyNums := regexp.MustCompile(`^\d+$`)
		port = os.Args[2]
		if !onlyNums.MatchString(port) {
			fmt.Fprint(os.Stderr, "incorrect port number provided")
			return 1
		}
	}

	// set development address, if needed
	if argNumber > 1 {
		lpath, ok := os.LookupEnv("LOCALPATH")
		if !ok {
			fmt.Fprint(os.Stderr, "no LOCALPATH defined in .env file")
			return 1
		}
		address = lpath + port
	}

	// copy static files
	staticFiles, err := os.ReadDir(staticDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read static directory: %s", err)
		return 1
	}
	for _, sf := range staticFiles {
		if sf.IsDir() {
			continue
		}
		err := os.Link(path(staticDir, sf.Name()), path(outputDir, sf.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't copy static files: %s", err)
			return 1
		}
	}

	// copy portraits
	err = os.Mkdir(path(outputDir, portraitDir), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create portrait directory: %s", err)
		return 1
	}
	portraitFiles, err := os.ReadDir(path(staticDir, portraitDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read portrait directory: %s", err)
		return 1
	}
	for _, pf := range portraitFiles {
		err := os.Link(path(staticDir, portraitDir, pf.Name()), path(outputDir, portraitDir, pf.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't copy portraits: %s", err)
			return 1
		}
	}

	// copy texts
	err = os.Mkdir(path(outputDir, textsDir), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create texts directory: %s", err)
		return 1
	}
	textsFiles, err := os.ReadDir(path(staticDir, textsDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read texts directory: %s", err)
		return 1
	}
	for _, tf := range textsFiles {
		err := os.Link(path(staticDir, textsDir, tf.Name()), path(outputDir, textsDir, tf.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't copy texts: %s", err)
			return 1
		}
	}

	// generate main page
	err = uniquePage("index.tmpl", funcMap, nil, path(outputDir, "index.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create index page: %s", err)
		return 1
	}

	// connect to db
	db, err := dbConnect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't connect to the database: %s", err)
		return 1
	}
	defer db.Close()

	// create tables and insert data
	err = db.Create("schema.sql", "data.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create database: %s", err)
		return 1
	}

	// generate authors list
	err = os.Mkdir(path(outputDir, authorsDir), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create authors directory: %s", err)
		return 1
	}
	authorsData, err := db.getAuthorData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get author data: %s", err)
		return 1
	}
	err = uniquePage("authorList.tmpl", funcMap, authorsData, path(outputDir, authorsDir, "index.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create authorList page: %s", err)
		return 1
	}

	// generate author pages
	authorTmpl := template.Must(template.New("author.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, "author.tmpl")))
	for _, a := range authorsData {
		// get data on author's works
		works, err := db.getAuthorWorks(a.Slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get author works data: %s", err)
			return 1
		}
		a.Works = works

		err = createPageInDir(a, authorTmpl, "authors")
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't create pages for authors: %s", err)
			return 1
		}
	}

	// generate works list
	err = os.Mkdir(path(outputDir, worksDir), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create works directory: %s", err)
		return 1
	}
	workData, err := db.getWorkData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get work data: %s", err)
		return 1
	}
	err = uniquePage("workList.tmpl", funcMap, workData, path(outputDir, worksDir, "index.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create workList page: %s", err)
		return 1
	}

	// generate work pages
	workTmpl := template.Must(template.New("work.tmpl").Funcs(funcMap).ParseFiles(path(templatesDir, "work.tmpl")))
	for _, w := range workData {
		// get work details
		authors, err := db.getWorkAuthors(w.Slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get work authors data: %s", err)
			return 1
		}
		w.Authors = authors

		editions, err := db.getWorkEditions(w.Slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get work editions data: %s", err)
			return 1
		}
		for i, e := range editions {
			links, err := db.getEditionLinks(e.Slug, e.Year)
			if err != nil {
				fmt.Fprintf(os.Stderr, "can't get edition link data: %s", err)
				return 1
			}
			editions[i].Links = links
		}
		w.Editions = editions

		translations, err := db.getWorkTranslations(w.Slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get work translations data: %s", err)
			return 1
		}
		for j, t := range translations {
			links, err := db.getEditionLinks(t.Slug, t.Year)
			if err != nil {
				fmt.Fprintf(os.Stderr, "can't get edition link data: %s", err)
				return 1
			}
			translations[j].Links = links
		}
		w.Translations = translations

		err = createPageInDir(w, workTmpl, "works")
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't create pages for works: %s", err)
			return 1
		}
	}

	// generate list of readable texts
	textsData, err := db.getTextsData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get texts data: %s", err)
		return 1
	}
	err = uniquePage("textsList.tmpl", funcMap, textsData, path(outputDir, textsDir, "index.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create readList page: %s", err)
		return 1
	}

	return 0
}

func path(files ...string) string {
	fileSlice := []string{}
	fileSlice = append(fileSlice, files...)
	return strings.Join(fileSlice, "/")
}
