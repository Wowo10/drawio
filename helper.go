package main

import (
	"fmt"
	"image"
	"image/draw"
	"unsafe"

	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
)

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

func drawFilledCircle(renderer *sdl.Renderer, cx, cy, radius int32) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				renderer.DrawPoint(cx+dx, cy+dy)
			}
		}
	}
}

func lerp(a, b float32, t float32) float32 {
	return a + (b-a)*t
}

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
