package poker

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"strings"
)

const JsonContentType = "application/json"

type PlayerStore interface {
	GetPlayerScore(name string) int
	RecordWin(name string)
	GetLeague() League
	DeletePlayer(name string)
}

type PlayerServer struct {
	store PlayerStore
	http.Handler
	template *template.Template
}

type Player struct {
	Name string
	Wins int
}

var (
	//go:embed "templates/*"
	gameTemplates embed.FS
)

const htmlTemplatePath = "templates/*.html"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewPlayerServer(store PlayerStore) (*PlayerServer, error) {
	p := new(PlayerServer)

	tmpl, err := template.ParseFS(gameTemplates, htmlTemplatePath)

	if err != nil {
		return nil, fmt.Errorf("problem opening %s %v", htmlTemplatePath, err)
	}

	p.template = tmpl
	p.store = store
	router := http.NewServeMux()
	router.Handle("/league", http.HandlerFunc(p.leagueHandler))
	router.Handle("/players/", http.HandlerFunc(p.playersHandler))
	router.Handle("/game", http.HandlerFunc(p.game))
	router.Handle("/ws", http.HandlerFunc(p.webSocket))
	p.Handler = router

	return p, nil
}

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", JsonContentType)
	err := json.NewEncoder(w).Encode(p.store.GetLeague())
	if err != nil {
		log.Printf("league handler encountered an error: %v", err)
	}
	w.WriteHeader(http.StatusOK)
}

func (p *PlayerServer) playersHandler(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")
	switch r.Method {
	case http.MethodPost:
		p.processWin(w, player)
	case http.MethodGet:
		p.showScore(w, player)
	case http.MethodDelete:
		p.deletePlayer(w, player)
	}
}

func (p *PlayerServer) game(w http.ResponseWriter, r *http.Request) {
	err := p.template.Execute(w, nil)
	if err != nil {
		log.Printf("template encountered an error: %v", err)
	}
}

func (p *PlayerServer) webSocket(w http.ResponseWriter, r *http.Request) {
	conn, _ := wsUpgrader.Upgrade(w, r, nil)
	_, winnerMessage, _ := conn.ReadMessage()
	p.store.RecordWin(string(winnerMessage))
}

func (p *PlayerServer) deletePlayer(w http.ResponseWriter, name string) {
	if p.store.GetPlayerScore(name) != 0 {
		p.store.DeletePlayer(name)
		w.WriteHeader(http.StatusOK)
	}
	w.WriteHeader(http.StatusNotFound)
}

func (p *PlayerServer) showScore(w http.ResponseWriter, player string) {
	score := p.store.GetPlayerScore(player)

	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter, player string) {
	p.store.RecordWin(player)
	w.WriteHeader(http.StatusAccepted)
}
