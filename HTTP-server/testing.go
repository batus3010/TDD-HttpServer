package poker

import (
	"testing"
	"time"
)

type StubPlayerStore struct {
	Scores   map[string]int
	WinCalls []string
	League   League
}

type ScheduledAlert struct {
	At     time.Duration
	Amount int
}

func (s *StubPlayerStore) DeletePlayer(name string) {
	// remove from Scores map
	delete(s.Scores, name)

	// filter out from League slice
	filtered := make(League, 0, len(s.League))
	for _, p := range s.League {
		if p.Name != name {
			filtered = append(filtered, p)
		}
	}
	s.League = filtered
}

func (s *StubPlayerStore) GetPlayerScore(name string) int {
	score := s.Scores[name]
	return score
}

func (s *StubPlayerStore) RecordWin(name string) {
	s.WinCalls = append(s.WinCalls, name)
}

func (s *StubPlayerStore) GetLeague() League {
	return s.League
}

func AssertPlayerWin(t testing.TB, store *StubPlayerStore, winner string) {
	t.Helper()

	if len(store.WinCalls) != 1 {
		t.Fatalf("got %d calls to RecordWin want %d", len(store.WinCalls), 1)
	}

	if store.WinCalls[0] != winner {
		t.Errorf("did not store correct winner got %q want %q", store.WinCalls[0], winner)
	}
}
