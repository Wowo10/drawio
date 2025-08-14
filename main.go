package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
)

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

	window.SetTitle("drawio")

	screenShotBackgroundTexture, err := imageToTexture(renderer, img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create texture: %v\n", err)
		os.Exit(1)
	}
	defer screenShotBackgroundTexture.Destroy()

	var (
		drawing      bool
		lastX, lastY int32
		currentShape Shape = Shape{
			color:     sdl.Color{R: 255, G: 0, B: 0, A: 255},
			brushSize: 4,
		}
	)

	canvas, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, mode.W, mode.H)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create canvas texture: %v\n", err)
		os.Exit(1)
	}
	canvas.SetBlendMode(sdl.BLENDMODE_BLEND)
	renderer.SetRenderTarget(canvas)
	renderer.SetDrawColor(0, 0, 0, 0)
	renderer.Clear()
	renderer.SetRenderTarget(nil)

	var undoStack []Shape

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

					if ctrlDown && zDown && len(undoStack) > 0 {
						undoStack = undoStack[:len(undoStack)-1]

						renderer.SetRenderTarget(canvas)
						renderer.SetDrawColor(0, 0, 0, 0)
						renderer.Clear()
						for _, s := range undoStack {
							s.Draw(renderer)
						}
						renderer.SetRenderTarget(nil)
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
						currentShape = dumpShape(renderer, canvas, currentShape, undoStack)
					}
				}
			case *sdl.MouseMotionEvent:
				if drawing {
					x, y := e.X, e.Y
					minDist := float32(currentShape.brushSize) * 0.8

					dx := float32(x - lastX)
					dy := float32(y - lastY)
					dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

					if dist >= minDist {
						steps := int(dist / minDist)
						if steps == 0 {
							steps = 1
						}

						for i := 1; i <= steps; i++ {
							t := float32(i) / float32(steps)
							px := int32(lerp(float32(lastX), float32(x), t))
							py := int32(lerp(float32(lastY), float32(y), t))
							currentShape.Add(px, py)

							if len(currentShape.Points) >= 8000 {
								currentShape = dumpShape(renderer, canvas, currentShape, undoStack)
							}
						}

						lastX, lastY = x, y
					}
				}
			}
		}

		renderer.Clear()
		renderer.Copy(screenShotBackgroundTexture, nil, nil)
		renderer.Copy(canvas, nil, nil)

		currentShape.Draw(renderer)

		mouseX, mouseY, _ := sdl.GetMouseState()
		drawFilledCircle(renderer, mouseX, mouseY, currentShape.brushSize)

		renderer.Present()

		sdl.Delay(16)
	}

	// prevent BadWindow error while SDL is closing
	time.Sleep(10 * time.Millisecond)
}

func dumpShape(renderer *sdl.Renderer, canvas *sdl.Texture, currentShape Shape, undoStack []Shape) Shape {
	renderer.SetRenderTarget(canvas)
	currentShape.Draw(renderer)
	renderer.SetRenderTarget(nil)

	undoStack = append(undoStack, currentShape)
	currentShape = Shape{color: currentShape.color, brushSize: currentShape.brushSize}
	return currentShape
}
