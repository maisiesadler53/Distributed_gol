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
	response := new(stubs.Response)
	client.Call(stubs.GenerateGameOfLife, request, response)

	newWorld := response.WorldPart
	turn <- response.Turn
	done <- true
	worldChan <- newWorld

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
thisLoop:
	for {
		select {
		case <-ticker.C:
			request := stubs.Request{}
			response := new(stubs.Response)
			client.Call(stubs.AliveCellCount, request, response)
			newWorld := response.WorldPart
			c.events <- AliveCellsCount{response.Turn, len(calculateAliveCells(p, newWorld))}
		case <-done:
			break thisLoop
		default: // If not, it continues
		}
	}
	fmt.Println("here")
	nextWorld = <-worldChan
	fmt.Println("There")
	turn = <-turnChan
	//for i := 0; i < p.Threads; i++ {
	//	part := <-worldParts[i]
	//	nextWorld = append(nextWorld, part...)
	//}
	world = append([][]byte{}, nextWorld...)
	nextWorld = [][]byte{}
	fmt.Println("There")
	// c.events <- TurnComplete{turn}
	// select {
	// case key := <-keyPresses:
	// 	if key == 's' {
	// 		c.ioCommand <- 0
	// 		filename = filename + "x" + strconv.Itoa(turn)
	// 		c.ioFilename <- filename
	// 		for i, row := range world {
	// 			for j := range row {
	// 				c.ioOutput <- world[i][j]
	// 			}
	// 		}
	// 		c.events <- ImageOutputComplete{turn, filename}
	// 	} else if key == 'q' {
	// 		c.ioCommand <- 0
	// 		filename = filename + "x" + strconv.Itoa(turn)
	// 		c.ioFilename <- filename
	// 		for i, row := range world {
	// 			for j := range row {
	// 				c.ioOutput <- world[i][j]
	// 			}
	// 		}

	// 		c.events <- ImageOutputComplete{turn, filename}
	// 		break executeLoop
	// 	} else if key == 'p' {
	// 		c.events <- StateChange{turn, Paused}

	// 		for {
	// 			keyAgain := <-keyPresses
	// 			if keyAgain == 'p' {
	// 				c.events <- StateChange{turn, Executing}
	// 				break
	// 			}
	// 		}
	// 	}
	// default:
	// }

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
	// ticker.Stop()

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
