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
)

func main() {
	// load environment vars
	err := godotenv.Load()
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

	// copy static files
	staticFiles, err := os.ReadDir(staticDir)
	if err != nil {
		panic(err)
	}
	for _, sf := range staticFiles {
		err := os.Link(path(staticDir, sf.Name()), path(outputDir, sf.Name()))
		if err != nil {
			panic(err)
		}
	}

	// generate main page
	err = uniquePage("index.tmpl", funcMap, nil, path(outputDir, "index.html"))
	if err != nil {
		panic(err)
	}

	// generate authors list
	err = os.Mkdir(path(outputDir, authorsDir), 0755)
	if err != nil {
		panic(err)
	}
	err = uniquePage("authorList.tmpl", funcMap, authorsData, path(outputDir, authorsDir, "index.html"))
	if err != nil {
		panic(err)
	}

	// generate author pages
	err = pageCollection("author.tmpl", funcMap, authorsMap, "authors")
	if err != nil {
		panic(err)
	}

	// generate works list
	err = os.Mkdir(path(outputDir, worksDir), 0755)
	if err != nil {
		panic(err)
	}
	err = uniquePage("workList.tmpl", funcMap, worksData, path(outputDir, worksDir, "index.html"))
	if err != nil {
		panic(err)
	}
}

func path(files ...string) string {
	fileSlice := []string{}
	fileSlice = append(fileSlice, files...)
	return strings.Join(fileSlice, "/")
}
