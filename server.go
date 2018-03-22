package netgame

import (
	"fmt"
	"time"
)

// Server holds the server-side game state
type Server struct {
	clients  []*Client
	entities []*Entity

	lastProcessedInput []int

	Network *LagNetwork

	updatesPerSecond float64
	tick             *time.Ticker

	renderer Renderer
}

func NewServer(renderer Renderer) *Server {
	s := &Server{
		Network: &LagNetwork{},
		renderer: renderer,
	}
	s.SetUpdateRate(10)
	go func() {
		for {
			<-s.tick.C
			s.Update()
		}
	} ()
	return s
}

func (s *Server) SetUpdateRate(hz float64) {
	s.updatesPerSecond = hz
	if s.tick != nil {
		s.tick.Stop()
	}
	s.tick = time.NewTicker(time.Second / time.Duration(hz))
}

func (s *Server) Connect(client *Client) {
	client.Server = s
	client.EntityId = len(s.clients)
	s.clients = append(s.clients, client)

	entity := &Entity{}
	s.entities = append(s.entities, entity)
	entity.ID = client.EntityId

	s.lastProcessedInput = append(s.lastProcessedInput, 0)

	if entity.ID == 0 {
		entity.X = 40
	} else {
		entity.X = 60
	}
}

func (s *Server) Update() {
	s.processInputs()
	s.sendWorldState()
	s.renderer.Render(s.entities)
}

func (s *Server) validateInput(input Input) bool {
	p := input.PressTime
	if p < 0 {
		p *= -1
	}
	return p <= 1./40. * 1000
}

func (s *Server) processInputs() {
	// process all pending messages from clients
	for {
		msg, found := s.Network.Recv()
		if !found {
			return
		}

		// Update the state of the entity, based on its input.
		// We just ignore inputs that don't look valid; this is what prevents clients from cheating.
		input := msg.(Input)
		if s.validateInput(input) {
			id := input.EntityId
			s.entities[id].ApplyInput(input)
			s.lastProcessedInput[id] = input.Sequence
		} else {
			fmt.Println("Input is invalid for entity", input.EntityId)
		}
	}
}

func (s *Server) sendWorldState() {
	var worldState WorldState
	for i := 0; i < len(s.clients); i++ {
		entity := s.entities[i]
		worldState.States = append(worldState.States, State{
			EntityId: entity.ID,
			X: entity.X,
			LastProcessedInput: s.lastProcessedInput[i],
		})
	}

	for _, client := range s.clients {
		client.Network.Send(client.LagMs, worldState)
	}
}
