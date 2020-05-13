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
	// TODO: Create a 2D slice to store the world.
	// TODO: Execute all turns of the Game of Life.
	// TODO: After each turn, send a slice of cells that are still alive to the c.alive channel.

	// Make sure that the IO has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.alive)
}
