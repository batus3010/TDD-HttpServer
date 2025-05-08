package main

import (
	poker "HTTP-server"
	"fmt"
	"log"
	"net/http"
)

const (
	dbFileName   = "game.db.json"
	localHostUrl = "http://localhost:5000"
)

func main() {
	store, closeFunc, err := poker.FileSystemPlayerStoreFromFile(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFunc()

	server, _ := poker.NewPlayerServer(store)
	fmt.Printf("listening on %s\n", localHostUrl)
	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}

}
