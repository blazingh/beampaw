package sshhandler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/blazingh/beampaw/helper"

	"github.com/gliderlabs/ssh"
	"github.com/pterm/pterm"
)

type Tunnel struct {
	FileName string
	FileSize int64
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
	// close the session when the client disconnect
	defer func() {
		s.Close()
	}()

	args, _ := helper.ParseArgs(s.Command())

	// handle help
	if _, ok := args["help"]; ok {
		handleHelp(s)
		return
	}

	// handle scp file streaming
	if _, ok := args["scp"]; ok {
		openSCPStream(s)
		return
	}

	// handle ssh file streaming
	openSSHStream(s)
	return

}

func handleHelp(s ssh.Session) {
	helper.PrintProjectHeader(s)
	pterm.DefaultBulletList.WithWriter(s).WithItems([]pterm.BulletListItem{
		// send a file
		{Level: 0, Text: "Send a file", TextStyle: pterm.NewStyle(pterm.FgCyan), BulletStyle: pterm.NewStyle(pterm.FgCyan)},
		{Level: 1, Text: "ssh beampaw.xyz < file.txt\n", TextStyle: pterm.NewStyle(pterm.FgLightWhite), Bullet: "$", BulletStyle: pterm.NewStyle(pterm.FgLightWhite)},

		// send a file with a specific name
		{Level: 0, Text: "Send a file with a specific name", TextStyle: pterm.NewStyle(pterm.FgCyan), BulletStyle: pterm.NewStyle(pterm.FgCyan)},
		{Level: 1, Text: "ssh beampaw.xyz name=myfile.txt < file.txt\n", TextStyle: pterm.NewStyle(pterm.FgLightWhite), Bullet: "$", BulletStyle: pterm.NewStyle(pterm.FgLightWhite)},

		// send a folder
		{Level: 0, Text: "Send a folder", TextStyle: pterm.NewStyle(pterm.FgCyan), BulletStyle: pterm.NewStyle(pterm.FgCyan)},
		{Level: 1, Text: "zip the folder", TextStyle: pterm.NewStyle(pterm.FgCyan), Bullet: "1-", BulletStyle: pterm.NewStyle(pterm.FgCyan)},
		{Level: 1, Text: "zip -r output.zip myfolder/", TextStyle: pterm.NewStyle(pterm.FgLightWhite), Bullet: "$", BulletStyle: pterm.NewStyle(pterm.FgLightWhite)},
		{Level: 1, Text: "send the zip file", TextStyle: pterm.NewStyle(pterm.FgCyan), Bullet: "2-", BulletStyle: pterm.NewStyle(pterm.FgCyan)},
		{Level: 1, Text: "ssh beampaw.xyz name=myfolder.zip < output.zip\n", TextStyle: pterm.NewStyle(pterm.FgLightWhite), Bullet: "$", BulletStyle: pterm.NewStyle(pterm.FgLightWhite)},
	}).Render()
	return
}

func openSSHStream(s ssh.Session) {

	helper.PrintProjectHeader(s)

	// parse args
	args, _ := helper.ParseArgs(s.Command())

	// get file name
	fileName, ok := args["name"]
	if !ok {
		pterm.Warning.WithWriter(s).Println("no file name provided, defaulting to 'file.txt'")
		fileName = "file.txt"
	}

	// create a new tunnel
	id := rand.Intn(math.MaxInt)
	tunnel := Tunnel{
		FileName: fileName,
		FileSize: 0,
		Writer:   make(chan io.Writer),
		DoneChan: make(chan struct{}),
	}
	OpenedTunnels[id] = tunnel

	defer func(id int) {
		delete(OpenedTunnels, id)
	}(id)

	// delete the file if the connection is closed
	go func(id int) {
		<-s.Context().Done()
		delete(OpenedTunnels, id)
		if _, ok := OpenedTunnels[id]; ok {
			fmt.Printf("%s : closing tunnel %d\n", s.User(), id)
		}
	}(id)

	// tunnel id
	pterm.Info.WithWriter(s).Println("tunnel id: " + strconv.Itoa(id))
	pterm.DefaultBasicText.WithWriter(s).Print("\n")

	// download page link
	pterm.DefaultBox.
		WithWriter(s).
		WithLeftPadding(2).
		WithRightPadding(2).
		WithTitle(pterm.FgCyan.Sprint("Download page")).
		Println(os.Getenv("WEB_URL") + "/download?id=" + strconv.Itoa(id))
	pterm.DefaultBasicText.WithWriter(s).Print("\n")

	// direct link
	pterm.DefaultBox.
		WithWriter(s).
		WithLeftPadding(2).
		WithRightPadding(2).
		WithTitle(pterm.FgCyan.Sprint("Direct download link")).
		Println(os.Getenv("WEB_URL") + "/file?id=" + strconv.Itoa(id))
	pterm.DefaultBasicText.WithWriter(s).Print("\n")

	loader1, _ := pterm.DefaultSpinner.WithWriter(s).Start("waiting for receiver...")

	fmt.Printf("%s : tunnel is ready: %d\n", s.User(), id)

	// close the tunnel when the connection is closed
	defer func() {
		close(tunnel.DoneChan)
	}()

	// wait for the writer from the http
	tunnelWriter := <-tunnel.Writer
	loader1.Info("Connection established")

	loader2, _ := pterm.DefaultSpinner.WithWriter(s).Start("sending file...")

	// send the file
	_, err := io.Copy(tunnelWriter, s)
	if err != nil {
		loader2.Fail("Error sending file")
	}

	loader2.Success("File sent")

	// close the tunnel and the connection
	fmt.Printf("%s : done sending file \n", s.User())
	return
}

func openSCPStream(s ssh.Session) {

	s.Write([]byte{0x00})

	reader := bufio.NewReader(s)
	header, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	headerItems := strings.Split(strings.Replace(header, "\n", " ", -1), " ")
	if len(headerItems) < 3 {
		return
	}

	// create a new tunnel
	id := rand.Intn(math.MaxInt)

	fmt.Printf("tunnel id: %s", strconv.Itoa(id))

	fileSizeInt, err := strconv.Atoi(headerItems[1])
	if err != nil {
		return
	}

	tunnel := Tunnel{
		FileName: headerItems[2],
		FileSize: int64(fileSizeInt),
		Writer:   make(chan io.Writer),
		DoneChan: make(chan struct{}),
	}
	OpenedTunnels[id] = tunnel

	defer func() {
		close(tunnel.DoneChan)
	}()

	s.Write([]byte{0x00})
	tunnelWriter := <-tunnel.Writer

	loader2, _ := pterm.DefaultSpinner.WithWriter(s).Start("sending file...")

	// send the file
	_, err = io.CopyN(tunnelWriter, s, int64(fileSizeInt))
	if err != nil {
		loader2.Fail("Error sending file")
	}

	s.Write([]byte{0x00})

	return
}
