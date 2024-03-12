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
	Code     string
	Language string
	Name     string
}

type work struct {
	Title       string
	Publication uint
}
