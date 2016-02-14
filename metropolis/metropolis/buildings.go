package metropolis

type Card interface {
}

/*

dieNum := userInput()

dieVal = rollDies(dieNum)
active = currentlyActivePlayer


// 1: check all red
// 2: check others


for _, player := range players {
	for _, b := range player.cards {
		if b.scope == others && active != player {
			b.Activate(player, active)
		}
	}
}

for _, player := range players {
	for _, b := range player.cards {
		if b.scope == any ||
			b.scope == myself && active == player {
				b.Activate(player, active)
			}
	}
}

toBuy := userInput()

if err := buy(active, toBuy); err != nil {
	send("cannot buy", err)
}

nextPlayer()


*/
