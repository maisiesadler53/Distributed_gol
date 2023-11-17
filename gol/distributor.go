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

//call
func callGenerateGameOfLife(client *rpc.Client, world [][]byte, params stubs.Params, startX int, endX int, startY int, endY int, quit chan bool, worldChan chan [][]byte, turn chan int, doneChan chan bool) {
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
	//once call is over tell the distributer to stop listening for commands and ticks
	//send turn and world to the distributer
	turn <- response.Turn
	worldChan <- response.WorldPart
	doneChan <- true
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
	//establish connection with RPC server and handle errors
	client, err := rpc.Dial("tcp", server)
	if err != nil {
		// Handle the error, e.g., log it or return
		fmt.Println("Error connecting to RPC server:", err)
		return
	}

	//close connection when distributer ends
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
			return
		}
	}(client)

	//read pgm image
	c.ioCommand <- 1 // command to read a pgm image
	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- filename // giving which file to read

	//create slice to store world
	world := make([][]byte, p.ImageWidth)
	nextWorld := [][]byte{}
	for i := range world {
		world[i] = make([]byte, p.ImageHeight)
	}

	//read pgm image into matrix
	for i, row := range world {
		for j := range row {
			world[i][j] = <-c.ioInput
			if world[i][j] == 255 {
				c.events <- CellFlipped{
					CompletedTurns: 0,
					Cell:           util.Cell{X: j, Y: i},
				}
			}
		}
	}

	//call the generateGameOfLife
	turn := 0
	quitChan := make(chan bool, 1)
	worldChan := make(chan [][]byte, 1)
	turnChan := make(chan int, 1)
	doneChan := make(chan bool, 1)
	go callGenerateGameOfLife(client, world, stubs.Params{Turns: p.Turns, Threads: p.Threads, ImageWidth: p.ImageHeight, ImageHeight: p.ImageWidth}, 0, p.ImageWidth, 0, p.ImageHeight, doneChan, worldChan, turnChan, doneChan)

	//listen for key presses or ticks until told to stop by the callGenerateGameOfLife function
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		quit := false

	tickerCtrlLoop:
		for {
			select {
			case <-ticker.C:
				//call AliceCellCount every tick, receive world and send to alivecell event
				request := stubs.Request{}
				response := new(stubs.Response)
				client.Call(stubs.AliveCellCount, request, response)
				newWorld := response.WorldPart
				c.events <- AliveCellsCount{response.Turn, len(calculateAliveCells(p, newWorld))}
			case key := <-keyPresses:
				if key == 's' {
					//call the Control rpc call and produce pgm image from the current world
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
					//tell the rpc to stop executing and leave the function loop
					ticker.Stop()
					quit = true
					request := stubs.Request{Ctrl: key}
					response := new(stubs.Response)
					client.Call(stubs.Control, request, response)
					break tickerCtrlLoop

				} else if key == 'p' {
					//call the control rpc and tell event to pause execution
					request := stubs.Request{Ctrl: key}
					response := new(stubs.Response)
					client.Call(stubs.Control, request, response)
					c.events <- StateChange{response.Turn, Paused}
					for {
						keyAgain := <-keyPresses
						if keyAgain == 'p' {
							//wait until p is pressed again to continue
							request := stubs.Request{Ctrl: key}
							response := new(stubs.Response)
							client.Call(stubs.Control, request, response)
							c.events <- StateChange{response.Turn, Executing}
							break
						}
					}
				} else if key == 'k' {
					//send k and break loop but don't tell to quit
					ticker.Stop()
					request := stubs.Request{Ctrl: key}
					response := new(stubs.Response)
					client.Call(stubs.Control, request, response)
					break tickerCtrlLoop
				}
			case <-doneChan:
				break tickerCtrlLoop
				//if the GenerateGameOfLife call ends then leave the loop
			default: // If no ticker or control continue
			}
		}
		quitChan <- quit
		return
	}()
	// receive the new world and number of turns from the generateGameOfLife call and put in a new
	quit := <-quitChan
	if quit {
		<-worldChan
		<-turnChan
		c.ioCommand <- ioCheckIdle
		<-c.ioIdle
		c.events <- StateChange{turn, Quitting}
		close(c.events)
	}

	nextWorld = <-worldChan
	turn = <-turnChan
	world = append([][]byte{}, nextWorld...)
	nextWorld = [][]byte{}

	//send matrix to make pgm
	c.ioCommand <- 0
	filename = filename + "x" + strconv.Itoa(turn)
	c.ioFilename <- filename
	for i, row := range world {
		for j := range row {
			c.ioOutput <- world[i][j] // Sending the matrix so the io can make a pgm
		}
	}
	c.events <- ImageOutputComplete{turn, filename}
	//report final state to events
	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.

	close(c.events)

}
