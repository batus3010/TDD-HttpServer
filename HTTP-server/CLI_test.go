package poker_test

import (
	poker "HTTP-server"
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

type GameSpy struct {
	StartedWith  int
	FinishedWith string
	StartCalled  bool
}

func (g *GameSpy) Start(numberOfPlayers int) {
	g.StartCalled = true
	g.StartedWith = numberOfPlayers
}

func (g *GameSpy) Finish(winner string) {
	g.FinishedWith = winner
}

type scheduledAlert struct {
	at     time.Duration
	amount int
}

func (s scheduledAlert) String() string {
	return fmt.Sprintf("%d chips at %v", s.amount, s.at)
}

type SpyBlindAlerter struct {
	alerts []scheduledAlert
}

func (s *SpyBlindAlerter) ScheduleAlertAt(at time.Duration, amount int) {
	s.alerts = append(s.alerts, scheduledAlert{at, amount})
}

var dummySpyAlerter = &SpyBlindAlerter{}
var dummyBlindAlerter = &SpyBlindAlerter{}
var dummyPlayerStore = &poker.StubPlayerStore{}
var dummyStdIn = &bytes.Buffer{}
var dummyStdOut = &bytes.Buffer{}

func TestCLI(t *testing.T) {
	t.Run("records a player name Chris win from user input", func(t *testing.T) {
		in := strings.NewReader("7\nChris wins\n")
		playerStore := &poker.StubPlayerStore{}
		game := poker.NewPokerGame(dummyBlindAlerter, playerStore)
		cli := poker.NewCLI(in, dummyStdOut, game)
		cli.PlayPoker()

		poker.AssertPlayerWin(t, playerStore, "Chris")
	})
	t.Run("records a player name Cleo win from user input", func(t *testing.T) {
		in := strings.NewReader("7\nCleo wins\n")
		playerStore := &poker.StubPlayerStore{}
		game := poker.NewPokerGame(dummyBlindAlerter, playerStore)
		cli := poker.NewCLI(in, dummyStdOut, game)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerStore, "Cleo")
	})
}

func TestGame_Start(t *testing.T) {
	t.Run("prints an error when a non numeric value is entered and does not start the game", func(t *testing.T) {
		stdOut := &bytes.Buffer{}
		in := strings.NewReader("Junk\n")
		game := &GameSpy{}

		cli := poker.NewCLI(in, stdOut, game)
		cli.PlayPoker()

		if game.StartCalled {
			t.Errorf("game should not have started")
		}

		gotPrompt := stdOut.String()
		wantPrompt := poker.StartGamePlayerPrompt + poker.BadPlayerInputErrMsg
		if gotPrompt != wantPrompt {
			t.Errorf("got %q, want %q", gotPrompt, wantPrompt)
		}
	})
	t.Run("schedule alerts on game on start for 5 players", func(t *testing.T) {
		blindAlerter := &SpyBlindAlerter{}
		game := poker.NewPokerGame(blindAlerter, dummyPlayerStore)
		game.Start(5)
		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{10 * time.Minute, 200},
			{20 * time.Minute, 300},
			{30 * time.Minute, 400},
			{40 * time.Minute, 500},
			{50 * time.Minute, 600},
			{60 * time.Minute, 800},
			{70 * time.Minute, 1000},
			{80 * time.Minute, 2000},
			{90 * time.Minute, 4000},
			{100 * time.Minute, 8000},
		}
		checkSchedulingCases(cases, t, blindAlerter)
	})

	t.Run("schedules alerts on game start for 7 players", func(t *testing.T) {
		blindAlerter := &SpyBlindAlerter{}
		game := poker.NewPokerGame(blindAlerter, dummyPlayerStore)

		game.Start(7)

		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{12 * time.Minute, 200},
			{24 * time.Minute, 300},
			{36 * time.Minute, 400},
		}

		checkSchedulingCases(cases, t, blindAlerter)
	})

	t.Run("it prompts the user to enter the number of players and starts the game", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		in := strings.NewReader("7\n")
		game := &GameSpy{}

		cli := poker.NewCLI(in, stdout, game)
		cli.PlayPoker()

		gotPrompt := stdout.String()
		wantPrompt := poker.StartGamePlayerPrompt

		if gotPrompt != wantPrompt {
			t.Errorf("got %q, want %q", gotPrompt, wantPrompt)
		}

		if game.StartedWith != 7 {
			t.Errorf("wanted Start called with 7 but got %d", game.StartedWith)
		}
	})
}

func TestGame_Finish(t *testing.T) {
	t.Run("finishes game with 'Chris' as winner", func(t *testing.T) {
		in := strings.NewReader("1\nChris wins\n")
		game := &GameSpy{}
		cli := poker.NewCLI(in, dummyStdOut, game)
		cli.PlayPoker()
		winner := "Chris"
		if game.FinishedWith != winner {
			t.Errorf("expected finish called with 'Chris' but got %q", game.FinishedWith)
		}
	})
}

func assertScheduledAlert(t *testing.T, got, want scheduledAlert) {
	t.Helper()
	if got.at != want.at {
		t.Errorf("scheduled alert at %v, want %v", got.at, want.at)
	}
}

func checkSchedulingCases(cases []scheduledAlert, t *testing.T, blindAlerter *SpyBlindAlerter) {
	t.Helper()
	for i, want := range cases {
		t.Run(fmt.Sprint(want), func(t *testing.T) {
			if len(blindAlerter.alerts) <= i {
				t.Fatalf("alert %d was not scheduled %v", i, blindAlerter.alerts)
			}
			got := blindAlerter.alerts[i]
			assertScheduledAlert(t, got, want)
		})
	}
}
