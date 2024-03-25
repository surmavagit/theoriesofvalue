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

	table, err := os.ReadFile("schema.sql")
	if err != nil {
		fmt.Printf("can't read file: %s", err)
		os.Exit(1)
	}
	_, err = db.Exec(string(table))
	if err != nil {
		fmt.Printf("can't create tables: %s", err)
		os.Exit(1)
	}

	data, err := os.ReadFile("data.sql")
	if err != nil {
		fmt.Printf("can't read file: %s", err)
		os.Exit(1)
	}
	_, err = db.Exec(string(data))
	if err != nil {
		fmt.Printf("can't fill database: %s", err)
		os.Exit(1)
	}

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
	authorsData, err := getAuthorData(db)
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
		works, err := getAuthorWorks(db, a.Slug)
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
	workData, err := getWorkData(db)
	if err != nil {
		panic(err)
	}
	err = uniquePage("workList.tmpl", funcMap, workData, path(outputDir, worksDir, "index.html"))
	if err != nil {
		panic(err)
	}
}

func path(files ...string) string {
	fileSlice := []string{}
	fileSlice = append(fileSlice, files...)
	return strings.Join(fileSlice, "/")
}
