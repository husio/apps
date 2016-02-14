package metropolis

import "container/ring"

type Player interface {
	ChooseDiesNum() int
	BuyCard() CardID
	Cards() []CardID
}

type CardID int

func Run(players []Player) {
	active := ring.New(len(players))
	for _, p := range players {
		active.Value = p
		active = active.Next()
	}

	for {
		active = active.Next()
		ap := active.Value.(Player)

		diesNum := ap.ChooseDiesNum()
		value := roll(diesNum)

		for _, p := range players {
			for _, cid := range player.Cards() {

			}
		}
	}
}

func roll(diesNum int) int {
	return 4
}
