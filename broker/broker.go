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

func (s *Broker) AliveCellCount(req stubs.Request, res *stubs.Response) (err error) {
	//tell GameOfLife that ticker has been sent
	s.tick <- true
	//return from function if the world and turn are received from generateGameOfLife
	res.WorldPart = <-s.world
	res.Turn = <-s.turn
	return

}

func (s *Broker) GenerateGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	fmt.Println("connected")
	var servers []string

	servers = append(servers, "52.87.222.194:8000")
	servers = append(servers, "52.204.81.92:8000")
	//servers = append(servers, "54.237.111.50:8000")
	//servers = append(servers, "100.26.236.219:8000")
	//servers = append(servers, "54.144.133.91:8000")
	//establish connection with RPC server and handle errors
	var clients []*rpc.Client
	for _, server := range servers {
		client, error := rpc.Dial("tcp", server)
		clients = append(clients, client)
		if error != nil {
			// Handle the error, e.g., log it or return
			fmt.Println("Error connecting to RPC server:", server, err)
			return
		}
	}
	fmt.Println("connected")
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

	p := req.Params
	turn := 0
	p.Threads = 2
	//make a world to contain the updated state each loop
	world := req.World
	nextWorld := [][]byte{}

	worldParts := make([]chan [][]byte, p.Threads)
	for i := range worldParts {
		worldParts[i] = make(chan [][]byte) // Channels for parallel calculation
	}

	//loop through each turn and update state
	//turnLoop:
	for turn = 0; turn < p.Turns; turn++ {
		//check if a key has been pressed or ticker
		//select {
		////if ticker received send world and turn
		//case <-s.tick:
		//	s.world <- world
		//	s.turn <- turn
		//case ctrl := <-s.ctrl:
		//	if ctrl == 's' {
		//		//if s control send the world and turn to the control function
		//		s.world <- world
		//		s.turn <- turn
		//	} else if ctrl == 'q' {
		//		s.world <- world
		//		s.turn <- turn
		//		//end the process by leaving the loop
		//		break turnLoop
		//	} else if ctrl == 'p' {
		//		//if p send world and wait in loop until p pressed again, then send again
		//		s.world <- world
		//		s.turn <- turn
		//		for {
		//			ctrlAgain := <-s.ctrl
		//			if ctrlAgain == 'p' {
		//				s.world <- world
		//				s.turn <- turn
		//				break
		//			}
		//		}
		//	} else if ctrl == 'k' {
		//		request := stubs.Request{}
		//		response := new(stubs.Response)
		//		for _, client := range clients {
		//			client.Call(stubs.Close, request, response)
		//		}
		//		s.world <- world
		//		s.turn <- turn
		//		s.closeListener <- true
		//		break turnLoop
		//	}
		//
		//	//if no ticker or ctrl just continue
		//default:
		//}
		//loop through the positions in the world and add up the number or surrounding live cells
		for i := 0; i < p.Threads; i++ {
			haloWorld := [][]byte{}

			haloWorld = append([][]byte{}, world[p.ImageHeight-1])
			haloWorld = append(haloWorld, world...)
			haloWorld = append(haloWorld, world[0])
			haloWorld = haloWorld[i*p.ImageHeight/p.Threads : (i+1)*p.ImageHeight/p.Threads+2]

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

		fmt.Println("about to append")
		for i := 0; i < p.Threads; i++ {
			fmt.Println("waiting for part")
			part := <-worldParts[i]
			fmt.Println("received part")
			nextWorld = append(nextWorld, part...)
		}

		fmt.Println("appended")
		//set the world to the nextWorld and reset the nextWorld
		world = nextWorld
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
