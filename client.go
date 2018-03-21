package netgame

import (
	"time"
)

type Client struct {
	// Local representation of the entities.
	Entities map[int]*Entity

	// Input state
	KeyLeft, KeyRight bool

	// Simulated network connection.
	Network *LagNetwork

	// This is a reference to the server so we can send data to it
	Server *Server

	// Latency
	LagMs float64

	// assigned by server on connect
	EntityId int

	UseClientSidePrediction bool
	UseServerReconciliation bool
	UseEntityInterpolation bool

	Sequence int
	PendingInputs []Input

	updateRate time.Duration
	tick       *time.Ticker

	lastTime float64

	renderer Renderer
}

type Renderer interface {
	Render([]*Entity)
}

func NewClient(renderer Renderer) *Client {
	c := &Client{
		EntityId: -1,
		Entities: map[int]*Entity{},
		Network: &LagNetwork{},
		renderer: renderer,
	}
	c.SetUpdateRate(time.Second/50)
	go func() {
		for {
			<-c.tick.C
			c.Update()
		}
	} ()
	return c
}

func (c *Client) SetUpdateRate(frequency time.Duration) {
	c.updateRate = frequency
	if c.tick != nil {
		c.tick.Stop()
	}
	c.tick = time.NewTicker(frequency)
}

func (c *Client) Update() {
	c.processServerMessages()

	if c.EntityId == -1 {
		return
	}

	c.processInputs()

	if c.UseEntityInterpolation {
		c.interpolateEntities()
	}

	var entities []*Entity
	for i := 0; i < len(c.Entities); i++ {
		entities = append(entities, c.Entities[i])
	}
	c.renderer.Render(entities)
}

func (c *Client) processInputs() {
	currentTime := timeNowMs()
	dt := currentTime - c.lastTime
	c.lastTime = currentTime

	var input Input
	if c.KeyRight {
		input.PressTime = dt
	} else if c.KeyLeft {
		input.PressTime = -dt
	} else {
		return
	}

	input.Sequence = c.Sequence
	c.Sequence++
	input.EntityId = c.EntityId
	c.Server.Network.Send(c.LagMs, input)

	if c.UseClientSidePrediction {
		c.Entities[c.EntityId].ApplyInput(input)
	}

	c.PendingInputs = append(c.PendingInputs, input)
}

func (c *Client) processServerMessages() {
	for {
		msg, found := c.Network.Recv()
		if !found {
			return
		}

		worldState := msg.(WorldState)
		for i := 0; i < len(worldState.States); i++ {
			state := worldState.States[i]

			// first time seeing this entity
			if len(c.Entities) <= state.EntityId {
				c.Entities[state.EntityId] = &Entity{ID: state.EntityId}
			}

			entity := c.Entities[state.EntityId]
			if entity.ID == c.EntityId {
				// Received the authoritative position of this client's entity.
				entity.X = state.X

				if c.UseServerReconciliation {
					// Re-apply all the inputs not yet processed by the server.
					for j := 0; j < len(c.PendingInputs); {
						input := c.PendingInputs[j]
						if input.Sequence <= state.LastProcessedInput {
							// Already processed. Its effect is already taken into account into the world update
							// we just got, so we can drop it.
							c.PendingInputs = c.PendingInputs[j+1:]
						} else {
							entity.ApplyInput(input)
							j++
						}
					}
				} else {
					// Reconciliation is disabled, so drop all the saved inputs.
					c.PendingInputs = c.PendingInputs[:]
				}
				continue
			}

			// Received the position of an entity other than this client's.
			if !c.UseEntityInterpolation {
				entity.X = state.X
			} else {
				entity.PositionBuffer = append(entity.PositionBuffer, PositionState{
					time: timeNowMs(),
					position: state.X,
				})
			}
		}
	}
}

func (c *Client) interpolateEntities() {
	now := timeNowMs()
	renderTimestamp := now - c.Server.updateRate

	for id, entity := range c.Entities {
		if id == c.EntityId {
			continue
		}

		buffer := entity.PositionBuffer

		// drop older positions
		for len(buffer) >= 2 && buffer[1].time <= renderTimestamp {
			buffer = buffer[1:]
		}

		// Interpolate between the two surrounding authoritative positions.
		if len(buffer) >= 2 && buffer[0].time <= renderTimestamp && renderTimestamp <= buffer[1].time {
			x0 := buffer[0].position
			x1 := buffer[1].position
			t0 := buffer[0].time
			t1 := buffer[1].time

			entity.X = x0 + (x1 - x0) * (renderTimestamp - t0) / (t1-t0)
		}
	}
}
