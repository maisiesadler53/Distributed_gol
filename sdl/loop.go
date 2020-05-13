package sdl

import (
	"github.com/veandco/go-sdl2/sdl"
	"time"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

func Start(p gol.Params, aliveCells <-chan []util.Cell, keyPresses chan<- rune) {
	w := NewWindow(int32(p.ImageWidth), int32(p.ImageHeight))

sdlLoop:
	for {
		event := w.PollEvent()
		if event != nil {
			switch event.(type) {
			case *sdl.KeyboardEvent:
				switch event.(*sdl.KeyboardEvent).Keysym.Sym {
				case sdl.K_p:
					keyPresses <- 'p'
				case sdl.K_s:
					keyPresses <- 's'
				case sdl.K_q:
					keyPresses <- 'q'
				}
			}
		}
		select {
		case cells, ok := <-aliveCells:
			if !ok {
				w.Destroy()
				break sdlLoop
			}
			w.ClearPixels()
			for _, c := range cells {
				w.SetPixel(c.X, c.Y)
			}
			w.RenderFrame()

		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

}
