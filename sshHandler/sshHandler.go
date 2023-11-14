package sshHandler

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

type Tunnel struct {
	FileName string
	Writer   chan io.Writer
	DoneChan chan struct{}
}

var OpenedTunnels = map[int]Tunnel{}

func Start() {
	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "2222"
	}

	ssh.Handle(handleConnection)

	fmt.Printf("SSH Listening on port %s\n", sshPort)
	log.Fatal(ssh.ListenAndServe(":"+sshPort, nil, ssh.HostKeyFile("./id_rsa")))
}

func handleConnection(s ssh.Session) {

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
	tunnel := Tunnel{
		FileName: nameCmd[1],
		Writer:   make(chan io.Writer),
		DoneChan: make(chan struct{}),
	}
	OpenedTunnels[id] = tunnel

	// delete the file when the connection is closed
	defer func(id int) {
		if _, ok := OpenedTunnels[id]; ok {
			delete(OpenedTunnels, id)
		}
	}(id)

	// delete the file if the connection is closed
	go func(id int) {
		<-s.Context().Done()
		if _, ok := OpenedTunnels[id]; ok {
			fmt.Printf("%s : closing tunnel %d\n", s.User(), id)
			delete(OpenedTunnels, id)
		}
	}(id)

	// header
	pterm.DefaultBasicText.WithWriter(s).Print("\n\n")
	pterm.DefaultBigText.
		WithWriter(s).
		WithLetters(
			putils.LettersFromStringWithStyle("BEAM ", pterm.FgCyan.ToStyle()),
			putils.LettersFromStringWithStyle("PAW", pterm.FgLightMagenta.ToStyle())).
		Render()
	pterm.DefaultBasicText.WithWriter(s).Print("\n\n")

	// link
	pterm.DefaultBox.
		WithWriter(s).
		WithLeftPadding(2).
		WithRightPadding(2).
		WithTitle(pterm.FgCyan.Sprint("direct download link")).
		Println("http://localhost:8888/file?id=" + strconv.Itoa(id))
	pterm.DefaultBasicText.WithWriter(s).Print("\n")

	loader1, _ := pterm.DefaultSpinner.WithWriter(s).Start("waiting for receiver...")

	fmt.Printf("%s : tunnel is ready: %d\n", s.User(), id)

	// close the tunnel when the connection is closed
	defer func() {
		close(tunnel.DoneChan)
	}()

	tunnelWriter := <-tunnel.Writer
	loader1.Info("Connection established")

	loader2, _ := pterm.DefaultSpinner.WithWriter(s).Start("sending file...")

	// send the file
	_, err := io.Copy(tunnelWriter, s)
	if err != nil {
		log.Fatal(err)
		loader2.Fail("Error sending file")
	}

	loader2.Success("File sent")

	// close the tunnel and the connection
	fmt.Printf("%s : done sending file \n", s.User())
}