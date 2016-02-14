package metropolis

type Player struct {
	Coins    int
	City     []Building
	IsActive bool
}

func NewPlayer() *Player {
	return &Player{
		Coins: 3,
		City:  []Building{},
		/*
				NewBuilding("Wheat Field", 0, All, []int{1}),
				NewBuilding("Bakery", 0, Myself, []int{2, 3}),
			},
		*/
	}
}

func (p *Player) Source(e Event) {
	switch ev := e.(type) {
	case *TurnStarted:
		p.IsActive = ev.Active == p
	case *DiceRolled:
		for _, b := range p.City {
			var sum int
			for _, die := range ev.Result {
				sum += die
			}
			for _, val := range b.ActValues() {
				if val == sum {
					b.Activate(sum, p, p)
					break
				}
			}
		}
	case *MoneyTransfered:
		if ev.Player == p {
			p.Coins = ev.Gained
		}
	case *BuildingBought:
		if ev.Turn.Active == p {
			p.City = append(p.City, ev.Building)
		}
	case *GameEnd:
	}
}
