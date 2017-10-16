package main

import (
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const BUFFERSIZE = 1024
const numberConnections = 8

var serverAddress string
var fileName string

func main() {
	fmt.Println(`
             *     ,MMM8&&&.            *
                  MMMM88&&&&&    .
                 MMMM88&&&&&&&
     *           MMM88&&&&&&&&
                 MMM88&&&&&&&&
                 'MMM88&&&&&&'
                   'MMM8&&&'      *
          |\___/|
          )     (             .              '
         =\     /=
           )===(       *
          /     \
          |     |
         /       \
         \       /
  _/\_/\_/\__  _/_/\_/\_/\_/\_/\_/\_/\_/\_/\_
  |  |  |  |( (  |  |  |  |  |  |  |  |  |  |
  |  |  |  | ) ) |  |  |  |  |  |  |  |  |  |
  |  |  |  |(_(  |  |  |  |  |  |  |  |  |  |
  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |  
`)
	if len(fileName) != 0 {
		runServer()
	} else if len(serverAddress) != 0 {
		runClient()
	}
}

// CLIENT CODE

func runClient() {
	uiprogress.Start()
	var wg sync.WaitGroup
	wg.Add(numberConnections)
	bars := make([]*uiprogress.Bar, numberConnections)
	for id := 0; id < numberConnections; id++ {
		go func(id int) {
			defer wg.Done()
			port := strconv.Itoa(27001 + id)
			connection, err := net.Dial("tcp", "localhost:"+port)
			if err != nil {
				panic(err)
			}
			defer connection.Close()

			bufferFileName := make([]byte, 64)
			bufferFileSize := make([]byte, 10)

			connection.Read(bufferFileSize)
			fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
			bars[id] = uiprogress.AddBar(int(fileSize+1028) / 1024).AppendCompleted().PrependElapsed()

			connection.Read(bufferFileName)
			fileName = strings.Trim(string(bufferFileName), ":")
			os.Remove(fileName + "." + strconv.Itoa(id))
			newFile, err := os.Create(fileName + "." + strconv.Itoa(id))
			if err != nil {
				panic(err)
			}
			defer newFile.Close()

			var receivedBytes int64
			for {
				if (fileSize - receivedBytes) < BUFFERSIZE {
					io.CopyN(newFile, connection, (fileSize - receivedBytes))
					// Empty the remaining bytes that we don't need from the network buffer
					connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
					break
				}
				io.CopyN(newFile, connection, BUFFERSIZE)
				//Increment the counter
				receivedBytes += BUFFERSIZE
				bars[id].Incr()
			}
		}(id)
	}
	wg.Wait()

	// cat the file
	os.Remove(fileName)
	finished, err := os.Create(fileName)
	defer finished.Close()
	if err != nil {
		log.Fatal(err)
	}
	for id := 0; id < numberConnections; id++ {
		fh, err := os.Open(fileName + "." + strconv.Itoa(id))
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(finished, fh)
		if err != nil {
			log.Fatal(err)
		}
		fh.Close()
		os.Remove(fileName + "." + strconv.Itoa(id))
	}
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("\n\n\nDownloaded " + fileName + "!")
	time.Sleep(1 * time.Second)
}

// SERVER CODE

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func runServer() {
	logger := log.WithFields(log.Fields{
		"function": "main",
	})
	logger.Info("Initializing")
	var wg sync.WaitGroup
	wg.Add(numberConnections)
	for id := 0; id < numberConnections; id++ {
		go listenerThread(id, &wg)
	}
	wg.Wait()
}

func listenerThread(id int, wg *sync.WaitGroup) {
	logger := log.WithFields(log.Fields{
		"function": "listenerThread@" + serverAddress + ":" + strconv.Itoa(27000+id),
	})

	defer wg.Done()

	err := listener(id)
	if err != nil {
		logger.Error(err)
	}
}

func listener(id int) (err error) {
	port := strconv.Itoa(27001 + id)
	logger := log.WithFields(log.Fields{
		"function": "listener@" + serverAddress + ":" + port,
	})
	server, err := net.Listen("tcp", serverAddress+":"+port)
	if err != nil {
		return errors.Wrap(err, "Error listening on "+serverAddress+":"+port)
	}
	defer server.Close()
	logger.Info("waiting for connections")
	//Spawn a new goroutine whenever a client connects
	for {
		connection, err := server.Accept()
		if err != nil {
			return errors.Wrap(err, "problem accepting connection")
		}
		logger.Info("Client connected")
		go sendFileToClient(id, connection)
	}
}

//This function is to 'fill'
func fillString(retunString string, toLength int) string {
	for {
		lengthString := len(retunString)
		if lengthString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

func sendFileToClient(id int, connection net.Conn) {
	logger := log.WithFields(log.Fields{
		"function": "sendFileToClient #" + strconv.Itoa(id),
	})
	defer connection.Close()
	//Open the file that needs to be send to the client
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	//Get the filename and filesize
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	numChunks := math.Ceil(float64(fileInfo.Size()) / float64(BUFFERSIZE))
	chunksPerWorker := int(math.Ceil(numChunks / float64(numberConnections)))

	bytesPerConnection := int64(chunksPerWorker * BUFFERSIZE)
	if id+1 == numberConnections {
		bytesPerConnection = fileInfo.Size() - (numberConnections-1)*bytesPerConnection
	}
	fileSize := fillString(strconv.FormatInt(int64(bytesPerConnection), 10), 10)

	fileName := fillString(fileInfo.Name(), 64)

	if id == 0 || id == numberConnections-1 {
		logger.Infof("numChunks: %v", numChunks)
		logger.Infof("chunksPerWorker: %v", chunksPerWorker)
		logger.Infof("bytesPerConnection: %v", bytesPerConnection)
		logger.Infof("fileName: %v", fileInfo.Name())
	}

	logger.Info("sending")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)

	chunkI := 0
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			//End of file reached, break out of for loop
			logger.Info("EOF")
			break
		}
		if (chunkI >= chunksPerWorker*id && chunkI < chunksPerWorker*id+chunksPerWorker) || (id == numberConnections-1 && chunkI >= chunksPerWorker*id) {
			connection.Write(sendBuffer)
		}
		chunkI++
	}
	fmt.Println("File has been sent, closing connection!")
	return
}
