package metropolis

import (
	"crypto/rand"
	"encoding/base32"
	"time"
)

type Event interface {
	ID() string
	Created() time.Time
}

type GameEvent struct {
	id      string
	created time.Time
}

func (e *GameEvent) ID() string {
	return e.id
}

func (e *GameEvent) Created() time.Time {
	return e.created
}

var (
	_ Event = &TurnStarted{}
	_ Event = &DiceRolled{}
	_ Event = &MoneyTransfered{}
	_ Event = &BuildingBought{}
	_ Event = &GameEnd{}
)

func initGameEvent(e *GameEvent) {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	s := base32.StdEncoding.EncodeToString(b[:])
	e.id = s[:8]
	e.created = time.Now()
}

type TurnStarted struct {
	GameEvent
	Active *Player
}

type DiceRolled struct {
	GameEvent
	Turn   *TurnStarted
	Result []int
}

type MoneyTransfered struct {
	GameEvent
	Turn   *TurnStarted
	Player *Player
	Gained int
}

type BuildingBought struct {
	GameEvent
	Turn     *TurnStarted
	Building Building
}

type GameEnd struct {
	GameEvent
	Winner *Player
}

type EventHub struct {
	stack []Event
	subs  map[EventConsumer]struct{}
}

type EventConsumer interface {
	Source(Event)
}

func NewEventHub() *EventHub {
	return &EventHub{
		subs: make(map[EventConsumer]struct{}),
	}
}

func (eh *EventHub) Publish(e Event) {
	for sub := range eh.subs {
		sub.Source(e)
	}
	eh.stack = append(eh.stack, e)
}
