package main

import (
	"log"
	"os"

	"github.com/blazingh/beampaw/httpHandler"
	"github.com/blazingh/beampaw/sshHandler"

	"github.com/joho/godotenv"
)

func init() {
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
}

func main() {

	go httpHandler.Start()

	sshHandler.Start()
}
