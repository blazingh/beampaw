package httphandler

import (
	"fmt"
	sshHandler "github.com/blazingh/beampaw/sshhandler"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

func Start() {

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8888"
	}

	fs := http.FileServer(http.Dir("./public"))

	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", handleLanding)

	http.HandleFunc("/file", handleFile)

	http.HandleFunc("/download", handleDownload)

	fmt.Printf("HTTP Listening on port %s\n", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}

func handleFile(w http.ResponseWriter, r *http.Request) {
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

func handleLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl, _ := template.New("base").ParseFiles("template/base.html", "template/home.html")
	err := tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
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

	data := struct {
		FileName    string
		DownloadURL string
	}{
		FileName:    openTunnel.FileName,
		DownloadURL: os.Getenv("WEB_URL") + "/file?id=" + idstr,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl, _ := template.New("base").ParseFiles("template/base.html", "template/download.html")
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}
