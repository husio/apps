package metropolis

import "testing"

func TestGame(t *testing.T) {
	p1 := NewPlayer()
	p2 := NewPlayer()

	customDeck := func() map[string][]*Building {
		return map[string][]*Building{
			"Test": clone(NewBuilding("Test Building", 2, Myself, []int{2, 4}), 4),
		}
	}

	g := NewGame([]*Player{p1, p2}, customDeck)

	diceRes := 3
	g.rollDice = func() int { return diceRes } // my lucky dice!

	_, err := g.Roll(p1, 2)
	if _, ok := err.(*InvalidActionError); !ok {
		t.Errorf("expected InvalidActionError, got %+v", err)
	}

}
