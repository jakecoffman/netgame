package netgame

import (
	"time"
)

// Msg is either an Input (client->server) or a WorldState (server->client)
type Msg interface{}

// Input is the client input (key press) to the server
type Input struct {
	PressTime float64
	Sequence int
	EntityId int
}

// WorldState is the server update to the clients
type WorldState struct {
	States []State
}

type State struct {
	EntityId int
	X float64
	LastProcessedInput int
}

// Record wraps the data with the time the data should be received, introducing latency
type Record struct {
	RecvTime float64
	Data Msg
}

// LagNetwork is just a list of the records that are queued up
type LagNetwork struct {
	Messages []Record
}

// Send records a message that will be received later
func (n *LagNetwork) Send(lagMs float64, message Msg) {
	n.Messages = append(n.Messages, Record{
		RecvTime: timeNowMs() + lagMs,
		Data: message,
	})
}

// Recv pulls a record off the queue if it's time, or returns false if there's none or it's not time
func (n *LagNetwork) Recv() (Msg, bool) {
	now := float64(time.Now().UnixNano() / int64(time.Millisecond))
	for i := 0; i < len(n.Messages); i++ {
		if n.Messages[i].RecvTime < now {
			data := n.Messages[i].Data
			n.Messages = n.Messages[i+1:]
			return data, true
		}
	}
	return nil, false
}

func timeNowMs() float64 {
	return float64(time.Now().UnixNano() / int64(time.Millisecond))
}
