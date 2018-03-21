package netgame

type Entity struct {
	ID int
	X float64
	Speed float64
	PositionBuffer []PositionState
}

type PositionState struct {
	time float64
	position float64
}

func (e *Entity) ApplyInput(input Input) {
	e.X += float64(input.PressTime) * 0.5
}
