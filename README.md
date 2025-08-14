# Drawio

A simple tool to draw on your screen during screenshare session.

After running it takes a screenshot of current active display (by mouse position) and renders a fullscreen window with it where you can draw using Left Mouse Button. To Close just click Escape button.

In other words the displays remains inactive / in draw mode. You can leave it pushing Escape button.

## Build
Build executable using below command:

```go build -tags static -o drawio```

It is required to have SDL2 installed in your OS.

```brew install sdl2```
or
```pacman -S sdl2```

and so on

## ShortCuts

* Esc - immediate exit
* +/- - adjust brush size
* r   - change color to Red
* g   - change color to Green
* b   - change color to Blue
* Ctrl + Z - undo
* Ctrl + S - save in home/drawio dir

###### Beershare License 🍺