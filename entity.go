package netgame

// Entity represents a game object
type Entity struct {
	ID int
	X float64
	Speed float64
	PositionBuffer []PositionState
}

// PositionState records where an entity was at a given time
type PositionState struct {
	time float64
	position float64
}

// ApplyInput is used by the server and by client interpolation to move the entity
func (e *Entity) ApplyInput(input Input) {
	e.X += float64(input.PressTime) * 0.1
}
