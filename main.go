package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	winWidth  = 800
	winHeight = 600
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Alien struct {
	Rect  *sdl.Rect
	Color sdl.Color
}

type Laser struct {
	Rect *sdl.Rect
	Dir  Direction
}

var (
	player    *sdl.Rect
	playerDir Direction
	aliens    []Alien
	lasers    []Laser
	running   bool
	alienLock sync.Mutex
	keys      = map[sdl.Keycode]bool{}
)

func initSDL() *sdl.Window {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalf("Could not initialize SDL: %s", err)
	}

	window, err := sdl.CreateWindow("Alien Invader Z", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatalf("Could not create window: %s", err)
	}

	return window
}

func initGame() {
	rand.Seed(time.Now().UnixNano())
	player = &sdl.Rect{X: 390, Y: 550, W: 20, H: 20}
	playerDir = Up
	addRandomAlien()
	running = true
}

func addRandomAlien() {
	x := int32(rand.Intn(winWidth - 20))
	y := int32(rand.Intn(winHeight - 20))
	colors := []sdl.Color{
		{R: 255, G: 0, B: 0, A: 255},   // Red
		{R: 165, G: 42, B: 42, A: 255}, // Brown
		{R: 0, G: 0, B: 255, A: 255},   // Blue
	}
	newAlien := Alien{Rect: &sdl.Rect{X: x, Y: y, W: 20, H: 20}, Color: colors[rand.Intn(len(colors))]}
	aliens = append(aliens, newAlien)
}

func checkCollision(a, b *sdl.Rect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func moveAliens() {
	for running {
		alienLock.Lock()
		for i, alien := range aliens {
			dx := int32(rand.Intn(5) - 2) // Random between -2 and 2
			dy := int32(rand.Intn(5) - 2) // Random between -2 and 2
			alien.Rect.X += dx
			alien.Rect.Y += dy
			// Make sure alien stays within window boundary
			if alien.Rect.X < 0 {
				alien.Rect.X = 0
			}
			if alien.Rect.Y < 0 {
				alien.Rect.Y = 0
			}
			if alien.Rect.X > winWidth-20 {
				alien.Rect.X = winWidth - 20
			}
			if alien.Rect.Y > winHeight-20 {
				alien.Rect.Y = winHeight - 20
			}
			aliens[i] = alien
		}
		alienLock.Unlock()
		time.Sleep(time.Millisecond * 50)
	}
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
	go moveAliens() // Start moving aliens in a separate goroutine

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				// Handle key events here
				keys[e.Keysym.Sym] = e.Type == sdl.KEYDOWN
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
		if keys[sdl.K_SPACE] {
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
		}

		for i, laser := range lasers {
			switch laser.Dir {
			case Up:
				laser.Rect.Y -= 5
			case Down:
				laser.Rect.Y += 5
			case Left:
				laser.Rect.X -= 5
			case Right:
				laser.Rect.X += 5
			}
			lasers[i] = laser

			for j, alien := range aliens {
				if checkCollision(laser.Rect, alien.Rect) {
					aliens = append(aliens[:j], aliens[j+1:]...)
					addRandomAlien()
					break
				}
			}
		}

		// Check if player collides with any alien
		for _, alien := range aliens {
			if checkCollision(player, alien.Rect) {
				fmt.Println("Game Over! You were touched by an alien!")
				running = false
				break
			}
		}
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.FillRect(player)

		for _, laser := range lasers {
			renderer.FillRect(laser.Rect)
		}

		alienLock.Lock()
		for _, alien := range aliens {
			renderer.SetDrawColor(alien.Color.R, alien.Color.G, alien.Color.B, alien.Color.A)
			renderer.FillRect(alien.Rect)
		}
		alienLock.Unlock()

		renderer.Present()

		time.Sleep(time.Millisecond * 16)
	}
	fmt.Println("Game Over!!!")
}
