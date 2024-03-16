package main

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
