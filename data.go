package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func (db DB) sqlQuery(columns []string, tables []string, rest string) (*sql.Rows, error) {
	return db.Query(fmt.Sprintf("SELECT %s FROM %s %s;", strings.Join(columns, ", "), strings.Join(tables, " "), rest))
}

func (db DB) selectFirstEditionYear() string {
	return "(SELECT MIN(year) AS year FROM edition WHERE edition.work_slug = work.slug)"
}

type Author struct {
	Slug        string
	Name        string
	Birth       *int
	Death       *int
	Wikidata    *string
	Wikipedia   *string
	OnlineBooks *string
	Works       []Work
}

type Work struct {
	Slug         string
	Page         bool
	Lang         string
	Year         *int
	AllAuthors   *string
	Authors      []Author
	Dubious      bool
	TitleMain    string
	FullTitle    *string
	Wikidata     *string
	Wikipedia    *string
	Editions     []Edition
	Translations []Translation
}

type Edition struct {
	Slug        string
	Year        int
	Important   bool
	Lang        string
	Description string
	Links       []Link
}

type Translation struct {
	Slug        string
	AllAuthors  *string
	Year        int
	Lang        string
	TitleMain   string
	Description string
	Links       []Link
}

type Link struct {
	Website     string
	Url         string
	Quality     string
	Download    bool
	Description *string
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
		return fmt.Errorf("can't read schema.sql file: %w", err)
	}
	_, err = db.Exec(string(table))
	if err != nil {
		return fmt.Errorf("can't create tables: %w", err)
	}

	data, err := os.ReadFile(dataFile)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no data.sql file found")
	}
	if err != nil {
		return fmt.Errorf("can't read data.sql file: %w", err)
	}
	_, err = db.Exec(string(data))
	if err != nil {
		return fmt.Errorf("can't fill database: %w", err)
	}

	return nil
}

func (db *DB) getAuthorData() ([]Author, error) {
	columns := []string{
		"slug",
		"birth",
		"death",
		"CONCAT_WS(' ', first_part, main_part, last_part) AS fullname",
		"wikidata",
		"CASE WHEN slug_pedia IS NOT NULL THEN CONCAT(lang_pedia", "'.wikipedia.org/wiki/'", "slug_pedia) END",
		"onlinebooks",
	}
	tables := []string{
		"author",
		fmt.Sprintf("INNER JOIN name ON author.slug = name.author AND name.lang = '%s'", siteLang),
		fmt.Sprintf("LEFT JOIN wikipedia ON author.wikidata = wikipedia.id AND wikipedia.site_lang = '%s'", siteLang),
	}
	rest := "WHERE page = true ORDER BY main_part"
	authorRows, err := db.sqlQuery(columns, tables, rest)
	if err != nil {
		return nil, err
	}
	defer authorRows.Close()

	authorData := []Author{}
	for authorRows.Next() {
		a := Author{}
		err := authorRows.Scan(&a.Slug, &a.Birth, &a.Death, &a.Name, &a.Wikidata, &a.Wikipedia, &a.OnlineBooks)
		if err != nil {
			return nil, err
		}
		authorData = append(authorData, a)
	}

	return authorData, nil
}

func (db *DB) getAuthorWorks(authorSlug string) ([]Work, error) {
	columns := []string{
		"slug",
		"page",
		"title.main_part",
		db.selectFirstEditionYear(),
	}
	tables := []string{
		"work",
		fmt.Sprintf("LEFT JOIN title ON work.slug = title.work_slug AND title.lang = '%s'", siteLang),
		"LEFT JOIN attribution ON work.slug = attribution.work_slug",
		fmt.Sprintf("LEFT JOIN name ON name.author = attribution.author_slug AND name.lang = '%s'", siteLang),
	}
	rest := fmt.Sprintf("WHERE name.author = '%s' ORDER BY year", authorSlug)
	workRows, err := db.sqlQuery(columns, tables, rest)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	authorWorks := []Work{}
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
	selectWorkAllAuthorsTable := fmt.Sprintf("(SELECT work_slug, STRING_AGG(name.main_part, ', ') AS names FROM attribution INNER JOIN name ON attribution.author_slug = name.author AND name.lang = '%s' GROUP BY work_slug)", siteLang)
	columns := []string{
		"authors.names",
		"INITCAP(eng_desc)",
		"slug",
		"wikidata",
		"CASE WHEN slug_pedia IS NOT NULL THEN CONCAT(lang_pedia, '.wikipedia.org/wiki/', slug_pedia) END",
		"title.main_part",
		"CASE WHEN title.first_part IS NOT NULL or title.last_part IS NOT NULL THEN CONCAT(title.first_part, title.main_part, title.last_part) END",
		db.selectFirstEditionYear(),
	}
	tables := []string{
		"work",
		fmt.Sprintf("INNER JOIN title ON title.work_slug = work.slug AND title.lang = '%s'", siteLang),
		fmt.Sprintf("LEFT JOIN %s AS authors ON work.slug = authors.work_slug", selectWorkAllAuthorsTable),
		fmt.Sprintf("LEFT JOIN wikipedia ON work.wikidata = wikipedia.id AND wikipedia.site_lang = '%s'", siteLang),
		"INNER JOIN lang ON work.lang = lang.three",
	}
	rest := "WHERE page = true ORDER BY year"
	workRows, err := db.sqlQuery(columns, tables, rest)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	for workRows.Next() {
		w := Work{}
		err := workRows.Scan(&w.AllAuthors, &w.Lang, &w.Slug, &w.Wikidata, &w.Wikipedia, &w.TitleMain, &w.FullTitle, &w.Year)
		if err != nil {
			return nil, err
		}
		workData = append(workData, w)
	}

	return workData, nil
}

func (db *DB) getWorkAuthors(workSlug string) ([]Author, error) {
	columns := []string{
		"author_slug",
		"CONCAT_WS(' ', first_part, main_part, last_part)",
	}
	tables := []string{
		"attribution",
		fmt.Sprintf("LEFT JOIN name ON attribution.author_slug = name.author AND name.lang = '%s'", siteLang),
	}
	rest := fmt.Sprintf("WHERE work_slug = '%s'", workSlug)
	rows, err := db.sqlQuery(columns, tables, rest)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workAuthors := []Author{}
	for rows.Next() {
		a := Author{}
		err := rows.Scan(&a.Slug, &a.Name)
		if err != nil {
			return nil, err
		}
		workAuthors = append(workAuthors, a)
	}

	return workAuthors, nil
}

func (db *DB) getWorkEditions(workSlug string) ([]Edition, error) {
	columns := []string{
		"work.slug",
		"year",
		fmt.Sprintf("CASE WHEN work.slug = '%s' THEN description ELSE CONCAT(INITCAP(lang.eng_desc), ' translation (', COALESCE((SELECT STRING_AGG(name.main_part, ', ') FROM attribution INNER JOIN name ON attribution.author_slug = name.author AND name.lang = '%s' WHERE attribution.work_slug = work.slug), 'anonymous'), '): \"', title.main_part, '\"') END", workSlug, siteLang),
	}
	tables := []string{
		"edition",
		"INNER JOIN work ON work.slug = edition.work_slug",
		"INNER JOIN lang ON lang.three = work.lang",
		"INNER JOIN title ON title.work_slug = work.slug",
	}
	rest := fmt.Sprintf("WHERE important = true AND (work.slug = '%s' OR translation = '%s')", workSlug, workSlug)
	rows, err := db.sqlQuery(columns, tables, rest)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	editions := []Edition{}
	for rows.Next() {
		e := Edition{}
		err := rows.Scan(&e.Slug, &e.Year, &e.Description)
		if err != nil {
			return nil, err
		}
		editions = append(editions, e)
	}

	return editions, nil
}

func (db *DB) getWorkTranslations(workSlug string) ([]Translation, error) {
	selectWorkAllAuthorsTable := fmt.Sprintf("(SELECT work_slug, STRING_AGG(name.main_part, ', ') AS names FROM attribution INNER JOIN name ON attribution.author_slug = name.author AND name.lang = '%s' GROUP BY work_slug)", siteLang)
	columns := []string{
		"work.slug",
		"authors.names",
		"year",
		"INITCAP(lang.eng_desc)",
		"title.main_part",
		"description",
	}
	tables := []string{
		"edition",
		fmt.Sprintf("INNER JOIN work ON work.slug = edition.work_slug AND translation = '%s'", workSlug),
		fmt.Sprintf("LEFT JOIN %s AS authors ON edition.work_slug = authors.work_slug", selectWorkAllAuthorsTable),
		"INNER JOIN lang ON lang.three = work.lang",
		"INNER JOIN title ON title.work_slug = work.slug",
	}
	rows, err := db.sqlQuery(columns, tables, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	translations := []Translation{}
	for rows.Next() {
		t := Translation{}
		err := rows.Scan(&t.Slug, &t.AllAuthors, &t.Year, &t.Lang, &t.TitleMain, &t.Description)
		if err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}

	return translations, nil
}
func (db *DB) getEditionLinks(workSlug string, editionYear int) ([]Link, error) {
	columns := []string{
		"website.label",
		"CONCAT(website.domain, website.url, source.url)",
		"source.quality",
		"source.download",
		"source.description",
	}
	tables := []string{
		"link_content",
		"INNER JOIN source ON link_content.sitename = source.sitename AND link_content.url = source.url",
		"INNER JOIN website ON source.sitename = website.sitename",
	}
	rest := fmt.Sprintf("WHERE work_slug = '%s' AND year = %d ORDER BY quality, download, website, length(description)", workSlug, editionYear)
	rows, err := db.sqlQuery(columns, tables, rest)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		l := Link{}
		err := rows.Scan(&l.Website, &l.Url, &l.Quality, &l.Download, &l.Description)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return links, nil
}
