package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	outputDir    = "public"
	staticDir    = "static"
	templatesDir = "templates"
	authorsDir   = "authors"
	worksDir     = "works"
	siteLang     = "eng"
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

	// copy static files
	staticFiles, err := os.ReadDir(staticDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read static directory: %s", err)
		return 1
	}
	for _, sf := range staticFiles {
		err := os.Link(path(staticDir, sf.Name()), path(outputDir, sf.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't copy static files: %s", err)
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

	for i, a := range authorsData {
		// get data on author's works
		works, err := db.getAuthorWorks(a.Slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't get author works data: %s", err)
			return 1
		}
		authorsData[i].Works = works
	}

	err = pageCollection("author.tmpl", funcMap, authorsData, "authors")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create authors pages: %s", err)
		return 1
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

	workPageData := []Work{}
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
			links, err := db.getEditionLinks(w.Slug, e.Year)
			if err != nil {
				fmt.Fprintf(os.Stderr, "can't get edition link data: %s", err)
				return 1
			}
			editions[i].Links = links
		}

		w.Editions = editions
		workPageData = append(workPageData, w)
	}

	err = pageCollection("work.tmpl", funcMap, workPageData, "works")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create works pages: %s", err)
		return 1
	}
	return 0
}

func path(files ...string) string {
	fileSlice := []string{}
	fileSlice = append(fileSlice, files...)
	return strings.Join(fileSlice, "/")
}
