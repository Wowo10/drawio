package main

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"unsafe"

	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		fmt.Fprintf(os.Stderr, "could not initialize sdl: %v\n", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	displayIndex, bounds, ss_bounds, mode, err := getDisplayForMouse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get display for mouse: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Opening on display %d at position %v with resolution %dx%d\n", displayIndex, bounds, mode.W, mode.H)

	fmt.Println(int(bounds.X), int(bounds.Y), int(bounds.W), int(bounds.H))

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
		points  []sdl.Point
		drawing bool
	)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			case *sdl.MouseButtonEvent:
				if t.Button == sdl.BUTTON_LEFT {
					switch t.Type {
					case sdl.MOUSEBUTTONDOWN:
						drawing = true
						points = append(points, sdl.Point{X: t.X, Y: t.Y})
					case sdl.MOUSEBUTTONUP:
						drawing = false
					}
				}
			case *sdl.MouseMotionEvent:
				if drawing {
					points = append(points, sdl.Point{X: t.X, Y: t.Y})
				}
			}
		}

		renderer.Clear()
		renderer.Copy(texture, nil, nil)

		renderer.SetDrawColor(255, 0, 0, 255)

		for _, p := range points {
			renderer.FillRect(&sdl.Rect{X: p.X, Y: p.Y, W: 6, H: 6})
		}

		renderer.Present()

		sdl.Delay(10)
	}
}

func getDisplayForMouse() (int, sdl.Rect, image.Rectangle, sdl.DisplayMode, error) {
	mouseX, mouseY, _ := sdl.GetGlobalMouseState()

	numDisplays, err := sdl.GetNumVideoDisplays()
	if err != nil {
		return -1, sdl.Rect{}, image.Rectangle{}, sdl.DisplayMode{}, err
	}

	for i := range numDisplays {
		bounds, err := sdl.GetDisplayBounds(i)
		if err != nil {
			return -1, sdl.Rect{}, image.Rectangle{}, sdl.DisplayMode{}, err
		}
		if mouseX >= bounds.X && mouseX < bounds.X+bounds.W &&
			mouseY >= bounds.Y && mouseY < bounds.Y+bounds.H {
			mode, err := sdl.GetDesktopDisplayMode(i)
			ss_bounds := screenshot.GetDisplayBounds(i)
			if err != nil {
				return i, bounds, ss_bounds, sdl.DisplayMode{}, err
			}
			return i, bounds, ss_bounds, mode, nil
		}
	}
	return -1, sdl.Rect{}, image.Rectangle{}, sdl.DisplayMode{}, fmt.Errorf("mouse not on any display")
}

func imageToTexture(renderer *sdl.Renderer, img image.Image) (*sdl.Texture, error) {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		bounds := img.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	}

	surface, err := sdl.CreateRGBSurfaceWithFormatFrom(
		unsafe.Pointer(&rgba.Pix[0]),
		int32(rgba.Rect.Dx()),
		int32(rgba.Rect.Dy()),
		32,
		int32(rgba.Stride),
		sdl.PIXELFORMAT_ABGR8888,
	)
	if err != nil {
		return nil, err
	}
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}
	return texture, nil
}
