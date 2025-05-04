package poker

import "time"

type PokerGame struct {
	alerter BlindAlerter
	store   PlayerStore
}

func NewPokerGame(alerter BlindAlerter, store PlayerStore) *PokerGame {
	return &PokerGame{
		alerter: alerter,
		store:   store,
	}
}

func (p *PokerGame) Start(numberOfPlayers int) {
	blindIncrement := time.Duration(5+numberOfPlayers) * time.Minute

	blinds := []int{100, 200, 300, 400, 500, 600, 800, 1000, 2000, 4000, 8000}
	blindTime := 0 * time.Second
	for _, blind := range blinds {
		p.alerter.ScheduleAlertAt(blindTime, blind)
		blindTime = blindTime + blindIncrement
	}
}

func (p *PokerGame) Finish(winner string) {
	p.store.RecordWin(winner)
}
