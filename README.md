# Build your own wormhole

## Server computer

For transfering `testfile` to others:

```
cd $GOPATH/src/github.com/schollz/wormhole
go build -ldflags "-s -w -X main.fileName=testfile" -o server.exe && ./server.exe
```

Also make sure to open up TCP ports `27001-27009` in your port forwarding.


## Client computer

Change `localhost` to the public address of the server computer:

```
cd $GOPATH/src/github.com/schollz/wormhole
go build -ldflags "-s -w -X main.serverAddress=localhost" -o client.exe
```