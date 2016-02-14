package metropolis

import (
	"fmt"
	"math/rand"
)

type Game struct {
	Players []*Player
	Piles   map[string][]Building

	active   *Player
	rollDice func() int
}

func NewGame(players []*Player, deck DeckBuilder) *Game {
	return &Game{
		Players: players,
		Piles:   deck(),
		active:  players[0], // 0 players is invalid
		rollDice: func() int {
			return rand.Intn(6) + 1
		},
	}
}

type DeckBuilder func() map[string][]Building

func (g *Game) Roll(p *Player, diceNum int) ([]int, error) {
	var (
		res []int
		err error
	)

	switch diceNum {
	case 2:
		has := false
		for _, b := range p.City {
			if b.Name() == "Station" {
				has = true
				break
			}
		}
		if !has {
			err = &InvalidActionError{
				desc: "cannot roll with two dice: Station not build",
			}
		} else {
			res = append(res, g.rollDice(), g.rollDice())
		}
	case 1:
		res = append(res, g.rollDice())
	default:
		err = &InvalidActionError{
			desc: "must roll with one or two dice",
		}
	}

	return res, err
}

func StdDeck() map[string][]Building {
	return map[string][]Building{}
	/*
			"Ranch":             clone(NewBuilding("Ranch", 1, All, []int{2}), 6),
			"Cafe":              clone(NewBuilding("Cafe", 2, Others, []int{3}), 6),
			"Convenience Store": clone(NewBuilding("Convenience Store", 2, Myself, []int{4}), 6),
		}
	*/
}

type InvalidActionError struct {
	desc string
}

func (e *InvalidActionError) Error() string {
	return fmt.Sprintf("invalid action: %s", e.desc)
}
