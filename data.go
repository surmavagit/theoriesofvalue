package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type Author struct {
	Slug        string
	Name        string
	Birth       *int
	Death       *int
	Wikidata    *string
	Wikipedia   *Wikipedia
	OnlineBooks *string
	Works       []Work
}

type Wikipedia struct {
	Code string
	Name string
}

type Work struct {
	Slug       string
	Page       bool
	Year       *int
	Authors    *string
	Dubious    bool
	TitleFirst *string
	TitleMain  string
	TitleLast  *string
	FullTitle  string
	Wikidata   *string
	Wikipedia  *Wikipedia
}

func dbConnect() (*DB, error) {
	host, err := getEnv("host")
	if err != nil {
		return nil, err
	}
	port, err := getEnv("port")
	if err != nil {
		return nil, err
	}
	user, err := getEnv("user")
	if err != nil {
		return nil, err
	}
	password, err := getEnv("password")
	if err != nil {
		return nil, err
	}
	dbname, err := getEnv("dbname")
	if err != nil {
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	return &DB{db}, err
}

func getEnv(envar string) (string, error) {
	result, ok := os.LookupEnv(envar)
	if !ok {
		return "", fmt.Errorf("no %s defined in .env file", envar)
	}
	return result, nil
}

func (db *DB) Create(schemaFile string, dataFile string) error {
	table, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("can't read file: %w", err)
	}
	_, err = db.Exec(string(table))
	if err != nil {
		return fmt.Errorf("can't create tables: %w", err)
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("can't read file: %w", err)
	}
	_, err = db.Exec(string(data))
	if err != nil {
		return fmt.Errorf("can't fill database: %w", err)
	}

	return nil
}

func (db *DB) getAuthorData() ([]Author, error) {
	query := "SELECT slug, birth, death, CONCAT(first_part, ' ', main_part, ' ', last_part) AS fullname, wikidata, onlinebooks FROM author INNER JOIN name ON author.slug = name.author AND name.lang = '" + siteLang + "'WHERE page = true ORDER BY main_part;"
	authorRows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer authorRows.Close()

	authorData := []Author{}
	for authorRows.Next() {
		a := Author{}
		err := authorRows.Scan(&a.Slug, &a.Birth, &a.Death, &a.Name, &a.Wikidata, &a.OnlineBooks)
		if err != nil {
			return nil, err
		}
		authorData = append(authorData, a)
	}

	return authorData, nil
}

func (db *DB) getAuthorWorks(authorSlug string) ([]Work, error) {
	authorWorks := []Work{}
	selectFirstEditionYear := "SELECT MIN(year) AS year FROM edition WHERE edition.work_slug = work.slug"
	query := "SELECT slug, page, title.main_part, (" + selectFirstEditionYear + ") FROM work LEFT JOIN title ON work.slug = title.work_slug AND title.lang = '" + siteLang + "' LEFT JOIN attribution ON work.slug = attribution.work_slug LEFT JOIN name ON name.author = attribution.author_slug AND name.lang = '" + siteLang + "' WHERE name.author = '" + authorSlug + "' ORDER BY year;"
	workRows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	for workRows.Next() {
		a := Work{}
		err := workRows.Scan(&a.Slug, &a.Page, &a.TitleMain, &a.Year)
		if err != nil {
			return nil, err
		}
		authorWorks = append(authorWorks, a)
	}

	return authorWorks, nil
}

func (db *DB) getWorkData() ([]Work, error) {
	workData := []Work{}
	selectFirstEditionYear := "SELECT MIN(year) AS year FROM edition WHERE edition.work_slug = work.slug"
	selectWorkAllAuthorsTable := "SELECT work_slug, STRING_AGG(name.main_part, ', ') AS names FROM attribution INNER JOIN name ON attribution.author_slug = name.author AND name.lang = '" + siteLang + "' GROUP BY work_slug"
	query := "SELECT authors.names, slug, dubious, wikidata, title.first_part, title.main_part, title.last_part, (" + selectFirstEditionYear + ") FROM work INNER JOIN title ON title.work_slug = work.slug AND title.lang = '" + siteLang + "' LEFT JOIN (" + selectWorkAllAuthorsTable + ") AS authors ON work.slug = authors.work_slug WHERE page = true;"
	workRows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	for workRows.Next() {
		w := Work{}
		err := workRows.Scan(&w.Authors, &w.Slug, &w.Dubious, &w.Wikidata, &w.TitleFirst, &w.TitleMain, &w.TitleLast, &w.Year)
		if err != nil {
			return nil, err
		}
		workData = append(workData, w)
	}

	return workData, nil
}
