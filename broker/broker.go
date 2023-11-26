package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
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

var ClientStates map[string]stubs.WorldState

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
	clientID := req.ID
	//make a world to contain the updated state each loop
	var world [][]byte
	var nextWorld [][]byte
	p := req.Params
	startTurn := 0
	turn := 0

	//if client previously connected (ID recognised) then use the state last saved for the client
	if state, exists := ClientStates[clientID]; exists {
		world = append([][]byte{}, state.World...)
		startTurn = state.Turn
	} else {
		world = append([][]byte{}, req.World...)
		nextWorld = [][]byte{}
	}

	worldParts := make([]chan [][]byte, p.Threads)
	for i := range worldParts {
		worldParts[i] = make(chan [][]byte) // Channels for parallel calculation
	}

	var servers []string
	for i := 0; i < p.Threads; i++ {
		servers = append(servers, "127.0.0.1:80"+strconv.Itoa(i)+"0")
	}

	//establish connection with RPC server and handle errors
	var clients []*rpc.Client
	for _, server := range servers {
		client, error := rpc.Dial("tcp", server)
		clients = append(clients, client)
		if error != nil {
			// Handle the error, e.g., log it or return
			fmt.Println("Error connecting to RPC server:", err)
			return
		}
	}

	//close connection when distributer ends
	for _, client := range clients {
		defer func(client *rpc.Client) {
			err := client.Close()
			if err != nil {
				fmt.Println("Error closing connection:", err)
				return
			}
		}(client)
	}

	//loop through each turn and update state
turnLoop:
	for turn = startTurn; turn < p.Turns; turn++ {
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
				turnsLeft := req.Params.Turns - turn
				req.Params.Turns = turnsLeft
				currentState := stubs.WorldState{
					World: world,
					Turn:  turn,
				}
				ClientStates[clientID] = currentState
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
				for _, client := range clients {
					client.Call(stubs.Close, request, response)
				}
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
			haloWorld := [][]byte{}
			if i == 0 {
				if p.Threads == 1 {
					haloWorld = append([][]byte{}, world[p.ImageHeight-1])
					haloWorld = append(haloWorld, world...)
					haloWorld = append(haloWorld, world[0])
				} else {
					haloWorld = append([][]byte{world[p.ImageHeight-1]}, world[:(i+1)*p.ImageHeight/p.Threads+2]...)
					fmt.Println(p.ImageHeight-1, ";", (i+1)*p.ImageHeight/p.Threads+2, ":first")
				}
			} else if i == (p.Threads - 1) {
				haloWorld = append(world[i*p.ImageHeight/p.Threads:], world[0])
				fmt.Println(i*p.ImageHeight/p.Threads, "; last")
			} else {
				haloWorld = world[i*p.ImageHeight/p.Threads : (i+1)*p.ImageHeight/p.Threads+2]
				fmt.Println(i*p.ImageHeight/p.Threads, ";", (i+1)*p.ImageHeight/p.Threads+2, ";middle")
			}
			fmt.Println(p.ImageHeight/p.Threads + 1)
			fmt.Println(len(haloWorld) - 2)
			req := stubs.Request{
				World:  haloWorld,
				Params: p,
				StartX: 1,
				EndX:   len(haloWorld) - 1,
				StartY: 0,
				EndY:   p.ImageWidth,
			}
			res := new(stubs.Response)
			go callWorker(clients[i], req, res, worldParts[i]) // every part goes to 1 worldPart channel
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
	ClientStates = make(map[string]stubs.WorldState)

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
