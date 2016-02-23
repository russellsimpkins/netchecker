DEFAULT: build

build:
	go build netchecker.go

dist:
	GOOS=linux GOARCH=amd64 go build netchecker.go

