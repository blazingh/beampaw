package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github/blazingh/go-stream/sshHandler"
)

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

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8888"
	}

	// start http server with a go routine
	go func() {
		// set up a file server
		fs := http.FileServer(http.Dir("./public"))
		http.Handle("/public/", http.StripPrefix("/public/", fs))

		http.HandleFunc("/home", handleHome)

		http.HandleFunc("/file", handleHttp)

		fmt.Printf("HTTP Listening on port %s\n", httpPort)
		log.Fatal(http.ListenAndServe(":"+httpPort, nil))
	}()

	sshHandler.Start()
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

	openTunnel, ok := sshHandler.OpenedTunnels[id]
	if !ok {
		w.Write([]byte("not found"))
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+openTunnel.FileName)

	// send when writer to the ssh
	openTunnel.Writer <- w
	// wait for the writer to close
	<-openTunnel.DoneChan
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFiles("template/home.html"))
	tmpl.Execute(w, nil)
}
