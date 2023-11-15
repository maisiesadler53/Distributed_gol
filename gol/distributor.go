package gol

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func makeCall(client *rpc.Client, world [][]byte, params stubs.Params, startX int, endX int, startY int, endY int, done chan bool, worldChan chan [][]byte, turn chan int) {
	request := stubs.Request{
		World:  world,
		Params: params,
		StartX: startX,
		EndX:   endX,
		StartY: startY,
		EndY:   endY,
	}
	//make response to hold the reply
	response := new(stubs.Response)
	//call GenerateGameOfLife
	client.Call(stubs.GenerateGameOfLife, request, response)
	//once call is over tell the function listening for commands and ticks to stop
	done <- true
	//send
	//newWorld := response.WorldPart
	turn <- response.Turn
	worldChan <- response.WorldPart
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	var cells []util.Cell
	for i, row := range world {
		for j, element := range row {
			if element == 255 {
				cells = append(cells, util.Cell{X: j, Y: i})
			}
		}
	}
	return cells
}

func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	server := "127.0.0.1:8040"
	client, err := rpc.Dial("tcp", server)
	if err != nil {
		// Handle the error, e.g., log it or return
		fmt.Println("Error connecting to RPC server:", err)
		return
	}
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
			return
		}
	}(client)

	c.ioCommand <- 1 // command to read a pgm image

	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- filename // giving which file to read

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageWidth)
	nextWorld := [][]byte{}
	for i := range world {
		world[i] = make([]byte, p.ImageHeight)
	}

	for i, row := range world {
		for j := range row {
			world[i][j] = <-c.ioInput // Writing the pgm image in a matrix
			if world[i][j] == 255 {
				c.events <- CellFlipped{
					CompletedTurns: 0,
					Cell:           util.Cell{X: j, Y: i},
				}
			}
		}
	}

	turn := 0

	// TODO: Execute all turns of the Game of Life.
	//worldParts := make([]chan [][]byte, p.Threads)
	//for i := range worldParts {
	//	worldParts[i] = make(chan [][]byte) // Channels for parallel calculation
	//}
	ticker := time.NewTicker(2 * time.Second) // This is for step 3
	done := make(chan bool, 1)
	worldChan := make(chan [][]byte, 1)
	turnChan := make(chan int, 1)

	go makeCall(client, world, stubs.Params{p.Turns, p.Threads, p.ImageHeight, p.ImageWidth}, 0, p.ImageWidth, 0, p.ImageHeight, done, worldChan, turnChan)
	go func() {
	ctrlTickerLoop:
		for {
			select {
			case <-ticker.C:
				fmt.Println("Ticker called")
				request := stubs.Request{}
				response := new(stubs.Response)
				client.Call(stubs.AliveCellCount, request, response)
				newWorld := response.WorldPart
				c.events <- AliveCellsCount{response.Turn, len(calculateAliveCells(p, newWorld))}
			case key := <-keyPresses:
				if key == 's' {
					request := stubs.Request{Ctrl: key}
					response := new(stubs.Response)
					client.Call(stubs.Control, request, response)
					c.ioCommand <- 0
					filename = filename + "x" + strconv.Itoa(response.Turn)
					c.ioFilename <- filename
					for i, row := range world {
						for j := range row {
							c.ioOutput <- response.WorldPart[i][j]
						}
					}
					c.events <- ImageOutputComplete{response.Turn, filename}
				} else if key == 'q' {
					request := stubs.Request{Ctrl: key}
					response := new(stubs.Response)
					client.Call(stubs.Control, request, response)
					return
				} else if key == 'p' {
					request := stubs.Request{Ctrl: key}
					response := new(stubs.Response)
					client.Call(stubs.Control, request, response)
					c.events <- StateChange{response.Turn, Paused}
					for {
						keyAgain := <-keyPresses
						if keyAgain == 'p' {
							request := stubs.Request{Ctrl: key}
							response := new(stubs.Response)
							client.Call(stubs.Control, request, response)

							c.events <- StateChange{response.Turn, Executing}
							break
						}
					}
				}
			case <-done:
				break ctrlTickerLoop
			default: // If not, it continues
			}
		}
	}()

	nextWorld = <-worldChan
	turn = <-turnChan
	world = append([][]byte{}, nextWorld...)
	nextWorld = [][]byte{}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}

	c.ioCommand <- 0
	filename = filename + "x" + strconv.Itoa(turn)
	c.ioFilename <- filename
	for i, row := range world {
		for j := range row {
			c.ioOutput <- world[i][j] // Sending the matrix so the io can make a pgm
		}
	}
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	ticker.Stop()

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
