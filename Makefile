run: build
	./relay 8080
build:
	go build -o relay main.go