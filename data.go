package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type author struct {
	Slug        string
	Name        string
	Birth       uint
	Death       uint
	WikidataId  string
	Wikipedia   wikipedia
	OnlineBooks string
	Works       []work
}

type wikipedia struct {
	Code string
	Name string
}

type work struct {
	Author      string
	Title       string
	Publication uint
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
