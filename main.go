package main

import (
	"fmt"
	"github.com/anthonynsimon/bild/blur"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
	"unsafe"
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

type Color int

const (
	Red Color = iota
	Brown
	Blue
)

func loadAndBlurTexture(renderer *sdl.Renderer, path string) (*sdl.Texture, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not load image: %v", err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("could not decode image: %v", err)
	}

	// Apply Gaussian blur using bild package
	blurredImg := blur.Gaussian(img, 3.0)

	// Convert image.Image to SDL2 surface
	surface, err := sdl.CreateRGBSurfaceFrom(unsafe.Pointer(&blurredImg.Pix[0]),
		int32(blurredImg.Rect.Dx()),
		int32(blurredImg.Rect.Dy()),
		32,
		blurredImg.Stride,
		0x000000FF,
		0x0000FF00,
		0x00FF0000,
		0xFF000000,
	)

	if err != nil {
		return nil, fmt.Errorf("could not create surface: %v", err)
	}
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, fmt.Errorf("could not create texture: %v", err)
	}

	return texture, nil
}
func loadMedia(renderer *sdl.Renderer) (*sdl.Texture, *mix.Music, error) {
	texture, err := loadAndBlurTexture(renderer, "level1.png")
	if err != nil {
		log.Fatal("Filed to load/blur image: %v", err)
	}
	if err := mix.Init(mix.INIT_MP3); err != nil {
		log.Fatalf("Initializing SDL_mixer failed: %v", err)
	}
	if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		log.Fatalf("Failed to open audio: %v", err)
	}

	music, err := mix.LoadMUS("level1.mp3")
	if err != nil {
		return nil, nil, fmt.Errorf("could not load background music: %v", err)
	}

	return texture, music, nil
}

func colorToSDLColor(color Color) sdl.Color {
	switch color {
	case Red:
		return sdl.Color{R: 255, G: 0, B: 0, A: 255}
	case Brown:
		return sdl.Color{R: 139, G: 69, B: 19, A: 255}
	case Blue:
		return sdl.Color{R: 0, G: 0, B: 255, A: 255}
	default:
		return sdl.Color{R: 255, G: 255, B: 255, A: 255}
	}
}

type Alien struct {
	Rect      *sdl.Rect
	FillColor Color
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
	color := Color(rand.Intn(3))
	newAlien := Alien{Rect: &sdl.Rect{X: x, Y: y, W: 20, H: 20}, FillColor: color}
	aliens = append(aliens, newAlien)
}

func createAliens(ticker *time.Ticker, alienCount int) {
	for range ticker.C {
		for i := 0; i < alienCount; i++ {
			addRandomAlien()
		}
	}
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
			switch alien.FillColor {
			case Red:
				if alien.Rect.X < player.X {
					alien.Rect.X += 5
				} else if alien.Rect.X > player.X {
					alien.Rect.X -= 5
				}
				if alien.Rect.Y < player.Y {
					alien.Rect.Y += 5
				} else if alien.Rect.Y > player.Y {
					alien.Rect.Y -= 5
				}
			default:

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
	texture, music, err := loadMedia(renderer)
	if err != nil {
		log.Fatalf("Could not load media: %s", err)
	}
	defer texture.Destroy()
	defer music.Free()

	music.Play(-1) // loop indefinitely
	initGame()
	// Create 5 aliens every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	go createAliens(ticker, 5)
	go moveAliens() // Start moving aliens in a separate goroutine

	for running {
		renderer.Copy(texture, nil, nil) // Copy the background image to fill the entire window

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
					//addRandomAlien()
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

		renderer.SetDrawColor(255, 255, 255, 255) // White color of player
		renderer.FillRect(player)

		renderer.SetDrawColor(0, 255, 0, 255) // Green color of laser
		for _, laser := range lasers {
			renderer.FillRect(laser.Rect)
		}

		alienLock.Lock()
		for _, alien := range aliens {
			sdlColor := colorToSDLColor(alien.FillColor)
			renderer.SetDrawColor(sdlColor.R, sdlColor.G, sdlColor.B, sdlColor.A)
			renderer.FillRect(alien.Rect)
		}
		alienLock.Unlock()

		renderer.Present()

		time.Sleep(time.Millisecond * 16)
	}
	fmt.Println("Game Over!!!")
}
