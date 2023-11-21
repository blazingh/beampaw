package httphandler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/blazingh/beampaw/helper"
	sshHandler "github.com/blazingh/beampaw/sshhandler"
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

	http.HandleFunc("/api/tunnel", handleGetTunnel)

	http.HandleFunc("/components/file", handleGetFile)

	fmt.Printf("HTTP Listening on port %s\n", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
func handleGetFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	idstr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid id"))
		return
	}

	openTunnel, ok := sshHandler.OpenedTunnels[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	splitFileName := strings.Split(openTunnel.FileName, ".")
	fileColor, ok := helper.FileExtensionsColors[splitFileName[len(splitFileName)-1]]
	if !ok {
		fileColor = "#000000"
	}

	data := struct {
		TunnelId      string `json:"tunnelId"`
		TunnelType    string `json:"tunnelType"`
		FileName      string `json:"fileName"`
		FileExtension string `json:"fileExtension"`
		FileColor     string `json:"fileColor"`
		DownloadURL   string `json:"downloadUrl"`
	}{
		TunnelId:      idstr,
		TunnelType:    "ssh",
		FileName:      openTunnel.FileName,
		FileExtension: splitFileName[len(splitFileName)-1],
		FileColor:     fileColor,
		DownloadURL:   os.Getenv("WEB_URL") + "/file?id=" + idstr,
	}

	tmpl, err := template.New("fileDownload").ParseFiles("template/components/fileDownload.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	err = tmpl.ExecuteTemplate(w, "fileDownload", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	return
}

func handleGetTunnel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	idstr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid id"))
		return
	}

	openTunnel, ok := sshHandler.OpenedTunnels[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	data := struct {
		TunnelId    string `json:"tunnelId"`
		TunnelType  string `json:"tunnelType"`
		FileName    string `json:"fileName"`
		DownloadURL string `json:"downloadUrl"`
	}{
		TunnelId:    idstr,
		TunnelType:  "ssh",
		FileName:    openTunnel.FileName,
		DownloadURL: os.Getenv("WEB_URL") + "/download?id=" + idstr,
	}

	jsonResp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)

	return
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

	return
}

func handleLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl, _ := template.New("base").ParseFiles("template/base.html", "template/home.html")
	err := tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, _ := template.New("base").ParseFiles("template/base.html", "template/download.html")
	err := tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
}
