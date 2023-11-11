package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/joho/godotenv"
)

type File struct {
	fileName string
	tunnel   chan Tunnel
}

type Tunnel struct {
	writer   io.Writer
	doneChan chan struct{}
}

var files = map[int]File{}

func main() {
	// check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		panic(".env file not found")
	}

	// check if id_rsa file exists
	if _, err := os.Stat("id_rsa"); os.IsNotExist(err) {
		panic("id_rsa file not found, you can run `ssh-keygen -t rsa -b 4096 -f id_rsa -q -N ''`")
	}

	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "2222"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8888"
	}

	// start http server with a go routine
	go func() {
		http.HandleFunc("/file", handleHttp)
		fmt.Printf("HTTP Listening on port %s\n", httpPort)
		log.Fatal(http.ListenAndServe(":"+httpPort, nil))
	}()

	ssh.Handle(hadleSsh)

	// start ssh server
	fmt.Printf("SSH Listening on port %s\n", sshPort)
	log.Fatal(ssh.ListenAndServe(":"+sshPort, nil, ssh.HostKeyFile("./id_rsa")))
}

func hadleSsh(s ssh.Session) {

	fmt.Printf("%s : session opened\n", s.User())

	// close the session when the client disconnect
	defer func() {
		s.Close()
	}()

	args := s.Command()
	if len(args) == 0 {
		s.Write([]byte("no arguments provided\n"))
		return
	}

	nameCmd := strings.Split(args[0], "=")
	if len(nameCmd) < 2 || nameCmd[0] != "name" || nameCmd[1] == "" {
		s.Write([]byte("invalid name command\n"))
		return
	}

	// create a new tunnel
	id := rand.Intn(math.MaxInt)
	file := File{
		fileName: nameCmd[1],
		tunnel:   make(chan Tunnel),
	}
	files[id] = file

	// delete the file when the connection is closed
	defer func(id int) {
		if _, ok := files[id]; ok {
			delete(files, id)
		}
	}(id)

	// delete the file if the connection is closed
	go func(id int) {
		<-s.Context().Done()
		if _, ok := files[id]; ok {
			fmt.Printf("%s : closing tunnel %d\n", s.User(), id)
			delete(files, id)
		}
	}(id)

	s.Write([]byte("tunnel id: " + strconv.Itoa(id) + "\n"))
	fmt.Printf("%s : tunnel is ready: %d\n", s.User(), id)

	// wait for the receiver to connect
	tunnel := <-file.tunnel
	// close the tunnel when the connection is closed
	defer func() {
		close(tunnel.doneChan)
	}()

	// send the file
	_, err := io.Copy(tunnel.writer, s)
	if err != nil {
		log.Fatal(err)
	}

	// close the tunnel and the connection
	s.Write([]byte("file received by receiver\n"))
	fmt.Printf("%s : done sending file \n", s.User())
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	idstr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	file, ok := files[id]
	if !ok {
		w.Write([]byte("not found"))
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+file.fileName)

	doneChan := make(chan struct{})
	file.tunnel <- Tunnel{
		writer:   w,
		doneChan: doneChan,
	}

	<-doneChan
}
