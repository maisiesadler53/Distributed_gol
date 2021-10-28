package main

import (
	"fmt"
	"time"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
)

var sdlEvents chan gol.Event
var sdlAlive chan int

func TestMain(m *testing.M) {
	p := gol.Params{ImageWidth: 512, ImageHeight: 512}
	sdlEvents = make(chan gol.Event)
	sdlAlive = make(chan int)
	result := make(chan int)
	go func() {
		result <- m.Run()
	}()
	// sdl.Run(p, sdlEvents, nil)
	w := sdl.NewWindow(int32(p.ImageWidth), int32(p.ImageHeight))

sdlLoop:
	for {
		w.PollEvent()
		select {
		case event, ok := <-sdlEvents:
			if !ok {
				w.Destroy()
				break sdlLoop
			}
			switch e := event.(type) {
			case gol.CellFlipped:
				w.FlipPixel(e.Cell.X, e.Cell.Y)
			case gol.TurnComplete:
				w.RenderFrame()
				sdlAlive <- w.CountPixels()
			case gol.FinalTurnComplete:
				w.Destroy()
				break sdlLoop
			default:
				if len(event.String()) > 0 {
					fmt.Printf("Completed Turns %-8v%v\n", event.GetCompletedTurns(), event)
				}
			}
		default:
			break
		}
	}
	os.Exit(<-result)
}

// TestSdl tests a 512x512 image for 100 turns using 8 worker threads.
func TestSdl(t *testing.T) {
	p := gol.Params{ImageWidth: 512, ImageHeight: 512, Turns: 100, Threads: 8}
	testName := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
	alive := readAliveCounts(p.ImageWidth, p.ImageHeight)
	t.Run(testName, func(t *testing.T) {
		turnNum := 0
		events := make(chan gol.Event)
		go gol.Run(p, events, nil)
		time.Sleep(2 * time.Second)
		final := false
		for event := range events {
			switch e := event.(type) {
				case gol.CellFlipped:
					sdlEvents <- e
				case gol.TurnComplete:
					turnNum++
					sdlEvents <- e
					aliveCount := <-sdlAlive
					if alive[turnNum] != aliveCount {
						t.Logf("Incorrect number of alive cells displayed on turn %d. Was %d, should be %d.", turnNum, aliveCount, alive[turnNum])
						time.Sleep(5 * time.Second)
						sdlEvents <- gol.FinalTurnComplete{}
						t.FailNow()
					} 
				case gol.FinalTurnComplete:
					final = true
					sdlEvents <- e
			}
		}

		if !final {
			sdlEvents <- gol.FinalTurnComplete{}
			t.Fatal("Simulation finished without sending a FinalTurnComplete event.")
		}
	})
}
