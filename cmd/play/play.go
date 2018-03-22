package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"image/color"
	"errors"
	"fmt"
	"os"
	"github.com/jakecoffman/netgame"
)

var (
	server *netgame.Server
	client1, client2 *netgame.Client

	serverRenderer = &renderer{}
	client1Renderer = &renderer{}
	client2Renderer = &renderer{}
)

var (
	ErrUserExit = errors.New("quitting")
)

func main() {
	server = netgame.NewServer(serverRenderer)
	client1 = netgame.NewClient(client1Renderer)
	client2 = netgame.NewClient(client2Renderer)
	server.Connect(client1)
	server.Connect(client2)

	// change these values to induce latency
	client1.LagMs = 250
	client2.LagMs = 150

	// immediately apply local changes to yourself to appear smoother
	client1.UseClientSidePrediction = true
	// when receiving position from server, playback unacknowledged inputs
	client1.UseServerReconciliation = true
	// view other entities slightly in the past so that their movements are smoother
	client1.UseEntityInterpolation = true

	err := ebiten.Run(update, 300, 240, 2, "Realtime Multiplayer")
	if err != nil && err != ErrUserExit {
		fmt.Fprintln(os.Stderr, err)
	}
}

func update(screen *ebiten.Image) error {
	// update
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ErrUserExit
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		client1.KeyRight = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyRight) {
		client1.KeyRight = false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		client1.KeyLeft = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyLeft) {
		client1.KeyLeft = false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		client2.KeyRight = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		client2.KeyRight = false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		client2.KeyLeft = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) {
		client2.KeyLeft = false
	}

	if ebiten.IsRunningSlowly() {
		return nil
	}

	// draw
	screen.Fill(color.Black)
	ebitenutil.DebugPrint(screen, fmt.Sprintf(`FPS: %0.2f
Client 1 Unacked: %v
Client 2 Unacked: %v

Client 1: (<- ->) %vms lag


Server:


Client 2: (A D) %vms lag
`, ebiten.CurrentFPS(), len(client1.PendingInputs), len(client2.PendingInputs), client1.LagMs, client2.LagMs))

	draw(screen, client1Renderer.entities, 90)
	draw(screen, serverRenderer.entities, 140)
	draw(screen, client2Renderer.entities, 190)
	return nil
}

func draw(screen *ebiten.Image, entities []*netgame.Entity, y float64) {
	for i, e := range entities {
		c := color.RGBA{B: 255, A: 255}
		if i == 1 {
			c = color.RGBA{R: 255, A: 255}
		}
		ebitenutil.DrawRect(screen, e.X, y, 10, 10, c)
	}
}

type renderer struct {
	entities []*netgame.Entity
}

func (r *renderer) Render(entities []*netgame.Entity) {
	r.entities = entities
}