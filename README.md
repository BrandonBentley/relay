# Relay

## Getting The Code
The Easiest way to download and run this code is to run the following command:

```bash
go get github.com/BrandonBentley/relay
```
Otherwise clone this repo to the following path:
```bash
$GOPATH/src/github.com/BrandonBentley/relay
```

## Go Version
This repo was successfully build using the following version of go:
```bash
1.12.7
```

## Build
The included Makefile has ever command need to run and test the relay server.
##### Build and Run Relay
```bash 
make
```
##### Build and Run Echo Server
```bash 
make echoserver
```

##### Build and Run Echo Client
```bash 
make echoclient
```

##### Build and Run Port Hog (attempts to occupy ports 8081-8099)
```bash 
make hog
```

## Programming Existing TCP Server to use Relay Service
If you are using Go you can use the simple implementation of the net.Listener interface found in [listener.go](tools\echoserver\relayclient\listener.go).

### Connection Procedure
1. Dial the Relay Service via TCP (example: localhost:8080)
   
2. After establishing a connection the port number to be sent back followed by a '\n' via the connection. (example: "8081\n")
   
   Note: This connection must remain live throughout the life of the relayed service.

3. Everytime a client connects via the relay port, the relay server will send a '\n' via the previously established connection. At that time create a new connection with the Relay Service the same way you created the inital connection in step 1 (example: localhost:8080)
   
4. As soon as the connection is established write the provided port number followed by a '\n' to the Connection (example: "8081\n")
   
5. The Relay Server will then connect the client and server connections together and communication will proceed as usual.