package main

import (
	"fmt"
	"time"
	// "io/ioutil"
	// "strconv"
	// "strings"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
	// "uk.ac.bris.cs/gameoflife/util"
)

var sdlEvents chan gol.Event

func TestMain(m *testing.M) {
	p := gol.Params{ImageWidth: 512, ImageHeight: 512}
	sdlEvents = make(chan gol.Event)
	go m.Run()
	sdl.Run(p, sdlEvents, nil)
}

// TestGol tests 16x16, 64x64 and 512x512 images on 0, 1 and 100 turns using 1-16 worker threads.
func TestSdl(t *testing.T) {
	p := gol.Params{ImageWidth: 16, ImageHeight: 16, Turns: 100, Threads: 8}
	testName := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
	t.Run(testName, func(t *testing.T) {
		events := make(chan gol.Event)
		go gol.Run(p, events, nil)
		time.Sleep(2 * time.Second)
		final := false
		for event := range events {
			switch e := event.(type) {
				case gol.CellFlipped:
					sdlEvents <- e
				case gol.TurnComplete:
					sdlEvents <- e
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
