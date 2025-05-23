package poker_test

import (
	poker "HTTP-server"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGETPlayers(t *testing.T) {
	store := poker.StubPlayerStore{
		Scores: map[string]int{
			"Pepper": 20,
			"Floyd":  10,
		},
	}
	server := mustMakePlayerServer(t, &store)
	t.Run("GET return Pepper's score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), "20")
	})

	t.Run("GET return Floyd's score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), "10")
	})
	t.Run("GET return 404 on missing player", func(t *testing.T) {
		request := newGetScoreRequest("Apollo")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func TestPostRecordWins(t *testing.T) {
	store := poker.StubPlayerStore{
		Scores: map[string]int{},
	}
	server := mustMakePlayerServer(t, &store)
	t.Run("it records wins when POST", func(t *testing.T) {
		player := "Pepper"
		request := newPostWinRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted) // return 202 Accepted on POST
		poker.AssertPlayerWin(t, &store, player)
	})
}

func TestDeletePlayer(t *testing.T) {
	store := poker.StubPlayerStore{
		Scores: map[string]int{
			"Pepper": 20,
			"Floyd":  10,
			"Batus":  5,
		},
	}
	server := mustMakePlayerServer(t, &store)
	t.Run("it returns 404 on non-existing player", func(t *testing.T) {
		player := "Peter"
		req, _ := http.NewRequest(http.MethodDelete, "/players/"+player, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatus(t, response.Code, http.StatusNotFound)
	})
	t.Run("it returns 200 on existing player and delete them", func(t *testing.T) {
		player := "Pepper"
		req, _ := http.NewRequest(http.MethodDelete, "/players/"+player, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatus(t, response.Code, http.StatusOK)
		req, _ = http.NewRequest(http.MethodDelete, "/players/Pepper", nil)
		response = httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func TestLeague(t *testing.T) {
	store := poker.StubPlayerStore{
		Scores: map[string]int{},
	}
	server := mustMakePlayerServer(t, &store)
	t.Run("it returns 200 on endpoint /League", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/league", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("it returns the League table as JSON", func(t *testing.T) {
		wantedLeague := []poker.Player{
			{"Cleo", 32},
			{"Chris", 20},
			{"Tits", 14},
		}
		store := poker.StubPlayerStore{nil, nil, wantedLeague}
		server := mustMakePlayerServer(t, &store)
		request := newLeagueRequest()
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		got := getLeagueFromResponse(t, response.Body)
		assertStatus(t, response.Code, http.StatusOK)
		assertContentType(t, response, poker.JsonContentType)
		assertLeague(t, got, wantedLeague)
	})
}

func TestWebGame(t *testing.T) {
	t.Run("GET /game returns status 200", func(t *testing.T) {
		server := mustMakePlayerServer(t, &poker.StubPlayerStore{})
		request := newGameRequest()
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
	})
	t.Run("message sent from websocket is the winner of the game", func(t *testing.T) {
		store := &poker.StubPlayerStore{}
		winner := "Cleo"
		server := httptest.NewServer(mustMakePlayerServer(t, store))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

		ws := mustDialWS(t, wsURL)
		defer func(ws *websocket.Conn) {
			err := ws.Close()
			if err != nil {
				t.Fatalf("could not close websocket connection on %s %v", server.URL, err)
			}
		}(ws)
		writeWSMessage(t, ws, winner)
		time.Sleep(10 * time.Millisecond)
		poker.AssertPlayerWin(t, store, winner)
	})

}

func assertLeague(t testing.TB, got, want []poker.Player) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func getLeagueFromResponse(t testing.TB, body io.Reader) (league []poker.Player) {
	t.Helper()
	//err := json.NewDecoder(body).Decode(&League)
	league, err := poker.NewLeague(body)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Player, '%v'", body, err)
	}
	return
}

func newLeagueRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/league", nil)
	return req
}

func newPostWinRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func newGetScoreRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/players/"+name, nil)
	return req
}

func newGameRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/game", nil)
	return req
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	if response.Result().Header.Get("content-type") != want {
		t.Errorf("response did not have content-type of %s, got %v", want, response.Result().Header)
	}
}

func mustMakePlayerServer(t *testing.T, store poker.PlayerStore) *poker.PlayerServer {
	server, err := poker.NewPlayerServer(store)
	if err != nil {
		t.Fatalf("could not create player server: %v", err)
	}
	return server
}

func mustDialWS(t *testing.T, url string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}

	return ws
}

func writeWSMessage(t testing.TB, conn *websocket.Conn, message string) {
	t.Helper()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		t.Fatalf("could not send message over ws connection %v", err)
	}
}
