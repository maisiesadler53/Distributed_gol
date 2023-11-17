package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
)

type Broker struct {
	tick          chan bool
	world         chan [][]byte
	turn          chan int
	ctrl          chan rune
	done          chan bool
	closeListener chan bool
}

func callWorker(client *rpc.Client, req stubs.Request, res *stubs.Response, worldChan chan [][]byte) {

	client.Call(stubs.GeneratePart, req, res)
	//once call is over tell the distributer to stop listening for commands and ticks
	//send turn and world to the distributer
	worldChan <- res.WorldPart
}

func (s *Broker) Control(req stubs.Request, res *stubs.Response) (err error) {
	//send control key to GenerateGameOfLife
	s.ctrl <- req.Ctrl
	//receive world from GenerateGameOflife and give to response
	res.WorldPart = <-s.world
	res.Turn = <-s.turn
	return
}

func (s *Broker) AliveCellCountTick(req stubs.Request, res *stubs.Response) (err error) {
	//tell GameOfLife that ticker has been sent
	s.tick <- true
	//return from function if the world and turn are received from generateGameOfLife
	res.WorldPart = <-s.world
	res.Turn = <-s.turn
	return

}

func (s *Broker) GenerateGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	fmt.Println("connected")
	server := "127.0.0.1:8030"
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

	//initiate variables
	// startX := req.StartX
	// startY := req.StartY
	// endX := req.EndX
	// endY := req.EndY
	// height := endY - startY
	// width := endX - startX
	p := req.Params
	turn := 0

	//make a world to contain the updated state each loop
	world := append([][]byte{}, req.World...)
	nextWorld := [][]byte{}

	worldParts := make([]chan [][]byte, p.Threads)
	for i := range worldParts {
		worldParts[i] = make(chan [][]byte) // Channels for parallel calculation
	}

	//loop through each turn and update state
turnLoop:
	for turn = 0; turn < p.Turns; turn++ {
		//check if a key has been pressed or ticker
		select {
		//if ticker received send world and turn
		case <-s.tick:
			s.world <- world
			s.turn <- turn
		case ctrl := <-s.ctrl:
			if ctrl == 's' {
				//if s control send the world and turn to the control function
				s.world <- world
				s.turn <- turn
			} else if ctrl == 'q' {
				s.world <- world
				s.turn <- turn
				//end the process by leaving the loop
				break turnLoop
			} else if ctrl == 'p' {
				//if p send world and wait in loop until p pressed again, then send again
				s.world <- world
				s.turn <- turn
				for {
					ctrlAgain := <-s.ctrl
					if ctrlAgain == 'p' {
						s.world <- world
						s.turn <- turn
						break
					}
				}
			} else if ctrl == 'k' {
				request := stubs.Request{}
				response := new(stubs.Response)
				client.Call(stubs.Close, request, response)
				fmt.Println("HERE")
				s.world <- world
				s.turn <- turn
				s.closeListener <- true
				break turnLoop
			}

			//if no ticker or ctrl just continue
		default:
		}
		//loop through the positions in the world and add up the number or surrounding live cells
		for i := 0; i < p.Threads; i++ {
			req := stubs.Request{
				World:  world,
				Params: p,
				StartX: i * p.ImageHeight / p.Threads,
				EndX:   (i + 1) * p.ImageHeight / p.Threads,
				StartY: 0,
				EndY:   p.ImageWidth,
			}
			res := new(stubs.Response)
			go callWorker(client, req, res, worldParts[i]) // every part goes to 1 worldPart channel
		}

		for i := 0; i < p.Threads; i++ {
			part := <-worldParts[i]
			nextWorld = append(nextWorld, part...)
		}
		//set the world to the nextWorld and reset the nextWorld
		world = append([][]byte{}, nextWorld...)
		nextWorld = [][]byte{}
	}
	//after all turns set the response to be the number of turns and the final world state
	res.WorldPart = world
	res.Turn = turn
	return
}

func main() {
	pAddr := flag.String("port", "8040", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	world := make(chan [][]byte)
	tick := make(chan bool)
	turn := make(chan int, 1)
	ctrl := make(chan rune)
	done := make(chan bool)
	closeListener := make(chan bool)

	//register rpc calls
	err := rpc.Register(&Broker{tick, world, turn, ctrl, done, closeListener})
	if err != nil {
		fmt.Println("Error registering listener", err)
		return
	}

	//listen on a TCP address
	listener, _ := net.Listen("tcp", ":"+*pAddr)

	//defer closing the listener: listener is closed when the function exits either
	//due to error or normal end
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener from defer", err)
			return
		}
	}(listener)

	//handles incoming RPC requests until closed
	go rpc.Accept(listener)
	<-closeListener
	return
}
