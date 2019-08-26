run: build
	./relay 8080

build:
	go build -o relay main.go

test:
	go test -timeout 30s -run TestEchoServer

hog:
	go build -o util/hog tools/portHog/hog.go
	./util/hog

echoserver: buildechoserver
	./util/echoserver localhost 8080

echoclient: buildechoclient
	./util/echoclient 8081

buildechoserver:
	go build -o util/echoserver tools/echoserver/echo.go

buildechoclient: 
	go build -o util/echoclient tools/echoclient/client.go
