package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
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

	// start a routine to handle http
	go func() {
		http.HandleFunc("/file", handleHttp)
		log.Fatal(http.ListenAndServe(":8888", nil))
	}()

	ssh.Handle(hadleSsh)

	// start ssh server
	log.Fatal(ssh.ListenAndServe(":2222", nil, ssh.HostKeyFile("./id_rsa")))
}

func hadleSsh(s ssh.Session) {
	fmt.Printf("session opened: %s\n", s.User())

	args := s.Command()
	if len(args) == 0 {
		s.Write([]byte("no arguments provided\n"))
		fmt.Printf("-- error : no arguments provided \n")
		return
	}

	nameCmd := strings.Split(args[0], "=")
	if len(nameCmd) < 2 || nameCmd[0] != "name" || nameCmd[1] == "" {
		s.Write([]byte("invalid name command\n"))
		fmt.Printf("-- error : invalid name command \n")
		return
	}

	// end function when the client disconnect
	go func() {
		<-s.Context().Done()
		fmt.Printf("-- message: session closed by client \n")
		return
	}()

	// create a new tunnel
	id := rand.Intn(math.MaxInt)
	file := File{
		fileName: nameCmd[1],
		tunnel:   make(chan Tunnel),
	}
	files[id] = file

	s.Write([]byte("tunnel id: " + strconv.Itoa(id) + "\n"))
	fmt.Printf("-- message: tunnel is ready: %d\n", id)

	// wait for the receiver to connect
	tunnel := <-file.tunnel

	// send the file
	_, err := io.Copy(tunnel.writer, s)
	if err != nil {
		log.Fatal(err)
	}

	// close the tunnel and the connection
	close(tunnel.doneChan)
	s.Write([]byte("file sent\n"))
	fmt.Printf("-- message: file sent \n")
	s.Close()
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
