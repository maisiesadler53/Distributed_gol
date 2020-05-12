package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	alive     chan<- []util.Cell
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	for i := 0; i < p.Turns; i++ {
		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				// Placeholder for the actual Game of Life logic: flips diagonal cells.
				if x == y {
					world[y][x] ^= 0xFF
				}
			}
		}

		// Create an empty slice to store coordinates of cells that are still alive after the turn.
		var aliveCells []util.Cell
		// Go through the world and append the cells that are still alive.
		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				if world[y][x] != 0 {
					aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
				}
			}
		}

		// Return the coordinates of cells that are still alive.
		c.alive <- aliveCells
	}

	// Make sure that the IO has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.alive)
}
