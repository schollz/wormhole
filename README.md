# Build your own wormhole

## Server computer

For transfering `testfile` to others:

```
cd $GOPATH/src/github.com/schollz/wormhole
go build -ldflags "-s -w -X main.fileName=testfile" -o server.exe && ./server.exe
```

Also make sure to open up TCP ports `27001-27009` in your port forwarding.

Then run:

```
./server.exe
```


## Client computer

Change `localhost` to the public address of the server computer:

```
cd $GOPATH/src/github.com/schollz/wormhole
go build -ldflags "-s -w -X main.serverAddress=localhost" -o client.exe
```

Then send the `client.exe` to whoever is going to recieve the file and have them double-click it.
