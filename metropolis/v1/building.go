package metropolis

type Building interface {
	Name() string
	Cost() int
	ActValues() []int
	Activate(roll int, active, buildingOwner *Player)
}

type building struct {
	name     string
	cost     int
	actVals  []int
	activate func(int, *Player, *Player)
}

func (b *building) Name() string {
	return b.name
}

func (b *building) Cost() int {
	return b.cost
}

func (b *building) ActValues() []int {
	return b.actVals
}

func (b *building) Activate(roll int, active, buildingOwner *Player) {
	b.activate(roll, active, buildingOwner)
}

func NewWheatField() Building {
	return &building{
		name:     "Wheat Field",
		cost:     1,
		activate: activateWheatField,
	}
}

func activateWheatField(roll int, active, owner *Player) {
	if roll == 1 {
		// owner gain 1
	}
}

func NewRanch() Building {
	return &building{
		name:     "Ranch",
		cost:     1,
		actVals:  []int{2},
		activate: activateRanch,
	}
}

func activateRanch(roll int, active, owner *Player) {
	// owner gain 1
}

func NewBakery() Building {
	return &building{
		name:     "Bakery",
		cost:     1,
		actVals:  []int{2, 3},
		activate: activateBakery,
	}
}

func activateBakery(roll int, active, owner *Player) {
	if active == owner {
		// owner gain 2
	}
}

func NewCafe() Building {
	return &building{
		name:     "Cafe",
		cost:     2,
		actVals:  []int{3},
		activate: activateBakery,
	}
}

func activateCafe(roll int, active, owner *Player) {
	if active != owner {
		// give owner 1 from active
	}
}
