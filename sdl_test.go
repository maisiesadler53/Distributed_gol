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

func TestMain(m *testing.M) {
	p := gol.Params{ImageWidth: 512, ImageHeight: 512}
	sdlEvents = make(chan gol.Event)
	result := make(chan int)
	go func() {
		result <- m.Run()
	}()
	sdl.Run(p, sdlEvents, nil)
	os.Exit(<-result)
}

// TestSdl tests a 512x512 image for 100 turns using 8 worker threads.
func TestSdl(t *testing.T) {
	p := gol.Params{ImageWidth: 512, ImageHeight: 512, Turns: 100, Threads: 8}
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
