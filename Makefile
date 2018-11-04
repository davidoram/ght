.PHONY:clean
clean:
	rm -f ./ght

.PHONY:build
build: clean go-deps
	go build .

.PHONY:install
install: clean go-deps
	go install .

.PHONY: go-deps
go-deps:
	go get golang.org/x/oauth2
	go get github.com/google/go-github/github

