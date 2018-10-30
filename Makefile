build: go-deps
	go build .

go-deps:
	go get golang.org/x/oauth2
	go get github.com/google/go-github/github