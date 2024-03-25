package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

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
	Slug     string
	Page     bool
	Authors  *string
	Dubious  bool
	Title    string
	Year     *int
	Wikidata *string
}

func dbConnect() (*sql.DB, error) {
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
	return sql.Open("postgres", psqlInfo)
}

func getEnv(envar string) (string, error) {
	result, ok := os.LookupEnv(envar)
	if !ok {
		return "", fmt.Errorf("no %s defined in .env file", envar)
	}
	return result, nil
}

func getAuthorData(db *sql.DB) ([]Author, error) {
	authorData := []Author{}
	query := "select slug, birth, death, CONCAT(first_part, ' ', main_part, ' ', last_part) as fullname, wikidata, onlinebooks from author right join name on author.slug = name.author where page = true order by main_part;"
	authorRows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer authorRows.Close()

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

func getAuthorWorks(db *sql.DB, slug string) ([]Work, error) {
	authorWorks := []Work{}
	query := fmt.Sprintf("select slug, page, title.main_part, (select MIN(year) as year from edition where edition.work = work.slug) from work left join title on work.slug = title.work left join attribution on work.slug = attribution.work left join name on name.author = attribution.author where name.author = '%s' order by year;", slug)
	workRows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	for workRows.Next() {
		a := Work{}
		err := workRows.Scan(&a.Slug, &a.Page, &a.Title, &a.Year)
		if err != nil {
			return nil, err
		}
		authorWorks = append(authorWorks, a)
	}

	return authorWorks, nil
}

func getWorkData(db *sql.DB) ([]Work, error) {
	workData := []Work{}
	selectShortTitle := " (select main_part from title where title.work = work.slug),"
	selectFirstEditionYear := " (select min(year) as year from edition where edition.work = work.slug)"
	selectWorkAllAuthorsTable := " (select work, string_agg((select main_part from name where name.author = attribution.author), ', ') as all_authors from attribution left join work on attribution.work = work.slug group by work)"
	query := "select page, dubious, wikidata, work, all_authors," + selectShortTitle + selectFirstEditionYear + " from work left join" + selectWorkAllAuthorsTable + " as authors on work.slug = authors.work;"
	workRows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	for workRows.Next() {
		w := Work{}
		err := workRows.Scan(&w.Page, &w.Dubious, &w.Wikidata, &w.Slug, &w.Authors, &w.Title, &w.Year)
		if err != nil {
			return nil, err
		}
		workData = append(workData, w)
	}

	return workData, nil
}
