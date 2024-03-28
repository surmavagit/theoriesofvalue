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
	// load environment vars
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// copy static files
	staticFiles, err := os.ReadDir(staticDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, sf := range staticFiles {
		err := os.Link(path(staticDir, sf.Name()), path(outputDir, sf.Name()))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// generate main page
	err = uniquePage("index.tmpl", funcMap, nil, path(outputDir, "index.html"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// connect to db
	db, err := dbConnect()
	if err != nil {
		fmt.Printf("can't connect to the database: %s", err)
		os.Exit(1)
	}
	defer db.Close() //panicking from now on

	// create tables and insert data
	err = db.Create("schema.sql", "data.sql")
	if err != nil {
		panic(err)
	}

	// generate authors list
	err = os.Mkdir(path(outputDir, authorsDir), 0755)
	if err != nil {
		panic(err)
	}
	authorsData, err := db.getAuthorData()
	if err != nil {
		panic(err)
	}
	err = uniquePage("authorList.tmpl", funcMap, authorsData, path(outputDir, authorsDir, "index.html"))
	if err != nil {
		panic(err)
	}

	// generate author pages

	for i, a := range authorsData {
		// get data on author's works
		works, err := db.getAuthorWorks(a.Slug)
		if err != nil {
			panic(err)
		}
		authorsData[i].Works = works
	}

	err = pageCollection("author.tmpl", funcMap, authorsData, "authors")
	if err != nil {
		panic(err)
	}

	// generate works list
	err = os.Mkdir(path(outputDir, worksDir), 0755)
	if err != nil {
		panic(err)
	}
	workData, err := db.getWorkData()
	if err != nil {
		panic(err)
	}
	err = uniquePage("workList.tmpl", funcMap, workData, path(outputDir, worksDir, "index.html"))
	if err != nil {
		panic(err)
	}

	workPageData := []Work{}
	for _, w := range workData {
		if !w.Page {
			continue
		}
		// get work details
		authors, err := db.getWorkAuthors(w.Slug)
		if err != nil {
			panic(err)
		}
		w.Authors = authors

		editions, err := db.getWorkEditions(w.Slug)
		if err != nil {
			panic(err)
		}
		for i, e := range editions {
			links, err := db.getEditionLinks(w.Slug, e.Year)
			if err != nil {
				panic(err)
			}
			editions[i].Links = links
		}

		w.Editions = editions
		workPageData = append(workPageData, w)
	}

	err = pageCollection("work.tmpl", funcMap, workPageData, "works")
	if err != nil {
		panic(err)
	}
}

func path(files ...string) string {
	fileSlice := []string{}
	fileSlice = append(fileSlice, files...)
	return strings.Join(fileSlice, "/")
}
