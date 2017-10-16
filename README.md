# Build your own wormhole

This program pays homage to [magic-wormhole](https://github.com/warner/magic-wormhole) except it doesn't have the rendevous server, or the transit relay, or the password-authenticated key exchange. Its not really anything like it, except that its file transfer over TCP. Here you can transfer a file using multiple TCP ports simultaneously. 

The binary has no flags even though a single binary has the client and server built-in. This is intentional. The flags are set at buildtime, so that each pair of server/client programs are used for only one type of file and can be deleted after the file is transfered. The binary knows whether it is a server or client depending on the flags set at buildtime. This is simpler for folks who don't know how to do anything except double-click on a program. Here, the computer-wizard will build both binaries and run the server on their computer. Then they send the client binary to the computerphobe who just double clicks on it and it will magically transfer the file straight to their computer, fast.

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
