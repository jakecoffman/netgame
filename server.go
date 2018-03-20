package netgame

import (
	"time"
	"fmt"
)

type Server struct {
	Clients []*Client
	Entities []*Entity

	lastProcessedInput []int

	Network LagNetwork

	updateRate float64
	tick       *time.Ticker

	renderer Renderer
}

func NewServer(renderer Renderer) *Server {
	s := &Server{renderer: renderer}
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
	s.updateRate = hz
	if s.tick != nil {
		s.tick.Stop()
	}
	s.tick = time.NewTicker(time.Second / time.Duration(hz))
}

func (s *Server) Connect(client *Client) {
	client.Server = s
	client.EntityId = len(s.Clients)
	s.Clients = append(s.Clients, client)

	entity := &Entity{}
	s.Entities = append(s.Entities, entity)
	entity.ID = client.EntityId

	s.lastProcessedInput = append(s.lastProcessedInput, 0)

	if entity.ID == 0 {
		entity.X = 4
	} else {
		entity.X = 6
	}
}

func (s *Server) Update() {
	s.processInputs()
	s.sendWorldState()
	s.renderer.Render(s.Entities)
}

func (s *Server) validateInput(input Input) bool {
	p := input.PressTime
	if p < 0 {
		p *= -1
	}
	return p <= 1/40 * 1000 * 1000
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
			s.Entities[id].ApplyInput(input)
			fmt.Println(input.Sequence)
			s.lastProcessedInput[id] = input.Sequence
		}
	}
}

func (s *Server) sendWorldState() {
	var worldState WorldState
	numClients := len(s.Clients)
	for i := 0; i < numClients; i++ {
		entity := s.Entities[i]
		worldState.States = append(worldState.States, State{
			EntityId: entity.ID,
			X: entity.X,
			LastProcessedInput: s.lastProcessedInput[i],
		})
	}

	for _, client := range s.Clients {
		client.Network.Send(client.LagMs, worldState)
	}
}
