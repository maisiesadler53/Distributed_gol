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

type GameOfLifeWorker struct {
	tick  chan bool
	world chan [][]byte
	turn  chan int
	ctrl  chan rune
	done  chan bool
}

func (s *GameOfLifeWorker) Control(req stubs.Request, res *stubs.Response) (err error) {
	//send control key to GenerateGameOfLife
	s.ctrl <- req.Ctrl
	fmt.Println("trying to get control")
	//receive world from GenerateGameOflife and give to response
	res.WorldPart = <-s.world
	res.Turn = <-s.turn
	fmt.Println("got the control")
	return
}

func (s *GameOfLifeWorker) AliveCellCountTick(req stubs.Request, res *stubs.Response) (err error) {
	//tell GameOfLife that ticker has been sent
	s.tick <- true
	//return from function if the world and turn are received from generateGameOfLife
	fmt.Println("trying to send count")
	res.WorldPart = <-s.world
	res.Turn = <-s.turn
	fmt.Println("got the count")
	return

}

func (s *GameOfLifeWorker) GenerateGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	//initiate variables
	startX := req.StartX
	startY := req.StartY
	endX := req.EndX
	endY := req.EndY
	height := endY - startY
	width := endX - startX
	p := req.Params
	turn := 0

	//make a world to contain the updated state each loop
	nextWorld := make([][]byte, width)
	for i := range nextWorld {
		nextWorld[i] = make([]byte, height)
	}

	//loop through each turn and update state
turnLoop:
	for turn = 0; turn < p.Turns; turn++ {
		fmt.Println("im looping")
		//check if a key has been pressed or ticker
		select {
		//if ticker received send world and turn
		case <-s.tick:
			s.world <- req.World
			s.turn <- turn
		case ctrl := <-s.ctrl:
			if ctrl == 's' {
				//if s control send the world and turn to the control function
				s.world <- req.World
				s.turn <- turn
			} else if ctrl == 'q' {
				//if q send world and turn and then end the process by leaving the loop
				s.world <- req.World
				s.turn <- turn
				break turnLoop
			} else if ctrl == 'p' {
				//if p send world and wait in loop until p pressed again, then send again
				s.world <- req.World
				s.turn <- turn
				for {
					ctrlAgain := <-s.ctrl
					if ctrlAgain == 'p' {
						s.world <- req.World
						s.turn <- turn
						break
					}
				}
			} else if ctrl == 'k' {
				s.world <- req.World
				s.turn <- turn
				break turnLoop
			}
			//if no ticker or ctrl just continue
		default:
		}
		//loop through the positions in the world and add up the number or surrounding live cells
		for i := startX; i < endX; i++ {
			for j := startY; j < endY; j++ {
				sum := 0
				adj := []int{-1, 0, 1}
				for _, n1 := range adj {
					for _, n2 := range adj {
						if n1 == 0 && n2 == 0 {
						} else if req.World[(i+n1+p.ImageWidth)%p.ImageWidth][(j+n2+p.ImageHeight)%p.ImageHeight] == 255 {
							sum++
						}
					}
				}
				//change cell depending on surrounding cells
				if (req.World[i][j] == 255) && (sum < 2 || sum > 3) {
					nextWorld[i-startX][j-startY] = 0
				} else if (req.World[i][j] == 0) && (sum == 3) {
					nextWorld[i-startX][j-startY] = 255
				} else {
					nextWorld[i-startX][j-startY] = req.World[i][j]
				}
			}
		}
		//set the world to the nextWorld and reset the nextWorld
		req.World = append([][]byte{}, nextWorld...)
		nextWorld = make([][]byte, width)
		for i := range nextWorld {
			nextWorld[i] = make([]byte, height)
		}
	}
	//after all turns set the response to be the number of turns and the final world state
	res.WorldPart = req.World
	res.Turn = turn
	fmt.Println("set the world ")
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

	//register rpc calls
	err := rpc.Register(&GameOfLifeWorker{tick, world, turn, ctrl, done})
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
			fmt.Println("Error closing listener", err)
			return
		}
	}(listener)

	//handles incoming RPC requests until closed
	rpc.Accept(listener)

}
