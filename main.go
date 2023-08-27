package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math/rand"
	"time"
)

const (
	winWidth  = 800
	winHeight = 600
)

type Direction int

const (
	shotCooldown           = 200 * time.Millisecond
	Up           Direction = iota
	Down
	Left
	Right
)

type Laser struct {
	Rect *sdl.Rect
	Dir  Direction
}

var (
	lastShot  time.Time
	player    *sdl.Rect
	playerDir Direction
	aliens    []*sdl.Rect
	lasers    []Laser
	running   bool
	keys      = map[sdl.Keycode]bool{}
)

func initSDL() *sdl.Window {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalf("Could not initialize SDL: %s", err)
	}

	window, err := sdl.CreateWindow("Alien Invader Z", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatalf("Could not create window: %s", err)
	}

	return window
}

func initGame() {
	rand.Seed(time.Now().UnixNano())
	player = &sdl.Rect{X: 390, Y: 550, W: 20, H: 20}
	playerDir = Up
	aliens = append(aliens, &sdl.Rect{X: 100, Y: 100, W: 20, H: 20})
	running = true
}

func addRandomAlien() {
	x := int32(rand.Intn(winWidth - 20))
	y := int32(rand.Intn(winHeight - 20))
	newAlien := &sdl.Rect{X: x, Y: y, W: 20, H: 20}
	aliens = append(aliens, newAlien)
}

func checkCollision(a, b *sdl.Rect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func main() {
	window := initSDL()
	defer window.Destroy()
	defer sdl.Quit()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("Could not create renderer: %s", err)
	}
	defer renderer.Destroy()

	initGame()

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				isKeyDown := e.Type == sdl.KEYDOWN
				keys[e.Keysym.Sym] = isKeyDown
			}
		}

		if keys[sdl.K_LEFT] && player.X > 0 {
			player.X -= 5
			playerDir = Left
		}
		if keys[sdl.K_RIGHT] && player.X < winWidth-20 {
			player.X += 5
			playerDir = Right
		}
		if keys[sdl.K_UP] && player.Y > 0 {
			player.Y -= 5
			playerDir = Up
		}
		if keys[sdl.K_DOWN] && player.Y < winHeight-20 {
			player.Y += 5
			playerDir = Down
		}
		if keys[sdl.K_SPACE] && time.Since(lastShot) >= shotCooldown {
			laser := Laser{}
			switch playerDir {
			case Up:
				laser.Rect = &sdl.Rect{X: player.X + 9, Y: player.Y, W: 2, H: 10}
			case Down:
				laser.Rect = &sdl.Rect{X: player.X + 9, Y: player.Y + 19, W: 2, H: 10}
			case Left:
				laser.Rect = &sdl.Rect{X: player.X, Y: player.Y + 9, W: 10, H: 2}
			case Right:
				laser.Rect = &sdl.Rect{X: player.X + 19, Y: player.Y + 9, W: 10, H: 2}
			}
			laser.Dir = playerDir
			lasers = append(lasers, laser)
			lastShot = time.Now()
		}
		// Updated laser movement logic
		for i := 0; i < len(lasers); i++ {
			laser := &lasers[i]
			switch laser.Dir {
			case Up:
				lasers[i].Rect.Y -= 5
			case Down:
				lasers[i].Rect.Y += 5
			case Left:
				lasers[i].Rect.X -= 5
			case Right:
				lasers[i].Rect.X += 5
			}

			// Existing alien collision logic
			for j, alien := range aliens {
				if checkCollision(lasers[i].Rect, alien) {
					aliens = append(aliens[:j], aliens[j+1:]...)
					lasers = append(lasers[:i], lasers[i+1:]...)
					addRandomAlien()
					break
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// Draw player
		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.FillRect(player)

		// Draw lasers
		renderer.SetDrawColor(0, 255, 0, 255)
		for _, laser := range lasers {
			renderer.FillRect(laser.Rect)
		}

		// Draw aliens
		renderer.SetDrawColor(255, 0, 0, 255)
		for _, alien := range aliens {
			renderer.FillRect(alien)
		}

		renderer.Present()
		sdl.Delay(16)
	}

	fmt.Println("Game Over!")
}