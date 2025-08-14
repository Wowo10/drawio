package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
)

type Shape struct {
	Points    []sdl.Point
	color     sdl.Color
	brushSize int32
}

func (s *Shape) Add(x, y int32) {
	s.Points = append(s.Points, sdl.Point{X: x, Y: y})
}

func (s Shape) Draw(r *sdl.Renderer) {
	r.SetDrawColor(s.color.R, s.color.G, s.color.B, s.color.A)
	for _, p := range s.Points {
		drawFilledCircle(r, p.X, p.Y, s.brushSize)
	}
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		fmt.Fprintf(os.Stderr, "could not initialize sdl: %v\n", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	_, bounds, ss_bounds, mode, err := getDisplayForMouse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get display for mouse: %v\n", err)
		os.Exit(1)
	}

	img, err := screenshot.Capture(ss_bounds.Min.X, ss_bounds.Min.Y, ss_bounds.Dx(), ss_bounds.Dy())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to capture screenshot: %v\n", err)
		os.Exit(1)
	}

	window, renderer, err := sdl.CreateWindowAndRenderer(
		mode.W, mode.H,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create window and renderer: %v\n", err)
		os.Exit(1)
	}

	window.SetPosition(bounds.X, bounds.Y)

	if err := window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set fullscreen: %v\n", err)
	}

	defer window.Destroy()
	defer renderer.Destroy()

	window.SetTitle("SDL2 Fullscreen on Mouse Display")

	texture, err := imageToTexture(renderer, img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create texture: %v\n", err)
		os.Exit(1)
	}
	defer texture.Destroy()

	var (
		shapes       []Shape
		drawing      bool
		lastX, lastY int32
		currentShape Shape = Shape{
			color:     sdl.Color{R: 255, G: 0, B: 0, A: 255},
			brushSize: 4,
		}
	)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					ctrlDown := e.Keysym.Mod&sdl.KMOD_CTRL != 0
					zDown := e.Keysym.Sym == sdl.K_z

					if ctrlDown && zDown {
						if len(shapes) != 0 {
							shapes = shapes[:len(shapes)-1]
						}
					}

					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
					case sdl.K_EQUALS:
						currentShape.brushSize++
					case sdl.K_MINUS:
						currentShape.brushSize--
					case sdl.K_r:
						currentShape.color = sdl.Color{R: 255, G: 0, B: 0, A: 255}
					case sdl.K_g:
						currentShape.color = sdl.Color{R: 0, G: 255, B: 0, A: 255}
					case sdl.K_b:
						currentShape.color = sdl.Color{R: 0, G: 0, B: 255, A: 255}
					}
				}

			case *sdl.MouseButtonEvent:
				if e.Button == sdl.BUTTON_LEFT {
					switch e.Type {
					case sdl.MOUSEBUTTONDOWN:
						drawing = true
						lastX, lastY = e.X, e.Y
					case sdl.MOUSEBUTTONUP:
						drawing = false
						shapes = append(shapes, currentShape)
						currentShape = Shape{color: currentShape.color, brushSize: currentShape.brushSize}
					}
				}
			case *sdl.MouseMotionEvent:
				if drawing {
					x, y := e.X, e.Y
					dx := float32(x - lastX)
					dy := float32(y - lastY)
					dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

					if dist == 0 {
						currentShape.Add(x, y)
					} else {
						step := float32(currentShape.brushSize) / 2
						for t := float32(0); t <= dist; t += step {
							px := int32(lerp(float32(lastX), float32(x), t/dist))
							py := int32(lerp(float32(lastY), float32(y), t/dist))
							currentShape.Add(px, py)
						}
					}

					lastX, lastY = x, y
				}
			}
		}

		renderer.Clear()
		renderer.Copy(texture, nil, nil)

		for _, s := range shapes {
			s.Draw(renderer)
		}

		currentShape.Draw(renderer)

		mouseX, mouseY, _ := sdl.GetMouseState()
		drawFilledCircle(renderer, mouseX, mouseY, currentShape.brushSize)

		renderer.Present()

		sdl.Delay(16)
	}

	// prevent BadWindow error while SDL is closing
	time.Sleep(10 * time.Millisecond)
}
