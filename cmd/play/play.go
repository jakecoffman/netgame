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
	"runtime"
)

var (
	server *netgame.Server
	client1, client2 *netgame.Client

	serverRenderer *renderer
	client1Renderer, client2Renderer *renderer
)

var (
	ErrUserExit = errors.New("quitting")
)

func main() {
	runtime.GOMAXPROCS(1)

	serverRenderer = &renderer{}
	client1Renderer = &renderer{}
	client2Renderer = &renderer{}

	server = netgame.NewServer(serverRenderer)
	client1 = netgame.NewClient(client1Renderer)
	client2 = netgame.NewClient(client2Renderer)
	server.Connect(client1)
	server.Connect(client2)

	client1.LagMs = 250
	client2.LagMs = 150

	client1.UseClientSidePrediction = true
	client1.UseEntityInterpolation = true
	client1.UseServerReconciliation = true

	err := ebiten.Run(update, 800, 600, 1, "Realtime Multiplayer")
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

	// draw
	screen.Fill(color.White)
	ebitenutil.DebugPrint(screen, "Hello world!")

	draw(screen, client1Renderer.entities, 100)
	draw(screen, serverRenderer.entities, 200)
	draw(screen, client2Renderer.entities, 300)
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