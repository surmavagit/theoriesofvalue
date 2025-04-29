all: serve

theoriesofvalue: $(wildcard *.go)
	go build ./...
data.sql:
	ln -sr ../theodata/data.sql data.sql
comments:
	ln -sr ../texts/comments comments
static/read:
	ln -sr ../texts/published static/read
public: theoriesofvalue data.sql comments static/read
	rm -rf public
	-dropdb theories
	createdb theories
	./theoriesofvalue -p
serve: public
	./ignored/serve


deploy: theoriesofvalue data.sql comments static/read
	rm -rf public
	-dropdb theories
	createdb theories
	./theoriesofvalue
	./ignored/deploy


clean:
	rm -rf public
	rm theoriesofvalue data.sql comments static/read
