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
	// done  bool // I am making this in case AliveCellsCountTick or Control sends smth through but after it gets stuck
}

func (s *GameOfLifeWorker) Control(req stubs.Request, res *stubs.Response) (err error) {
	fmt.Println("function called")
	s.ctrl <- req.Ctrl
	res.WorldPart = <-s.world
	res.Turn = <-s.turn
	return
}

func (s *GameOfLifeWorker) AliveCellCountTick(req stubs.Request, res *stubs.Response) (err error) {
	// res.WorldPart = s.world
	fmt.Println("alive function called")
	s.tick <- true
	// if s.done {
	// 	return
	// }
loop:
	for {
		select {
		case <-s.done:
			return
		// case worldPart := <-s.world:
		// 	res.WorldPart = worldPart
		default:
			res.WorldPart = <-s.world
			res.Turn = <-s.turn
			break loop
		}
	}

	fmt.Println(" alive function finished")
	return
}

func (s *GameOfLifeWorker) GenerateGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	fmt.Println("Got Message ")
	startX := req.StartX
	startY := req.StartY
	endX := req.EndX
	endY := req.EndY
	height := endY - startY
	width := endX - startX
	p := req.Params
	world := make([][]byte, p.ImageWidth)

	for i := range world {
		world[i] = make([]byte, p.ImageHeight)
	}
	world = append([][]byte{}, req.World...)
	nextWorld := make([][]byte, width)
	for i := range nextWorld {
		nextWorld[i] = make([]byte, height)
	}
	turn := 0
turnLoop:
	for turn = 0; turn < p.Turns; turn++ {
		// fmt.Println("Keeps running")
		select {
		case <-s.tick:
			fmt.Println("inside for tick")
			s.world <- world
			s.turn <- turn
			fmt.Println("sent world")
		case ctrl := <-s.ctrl:
			if ctrl == 's' {
				s.world <- world
				s.turn <- turn
			} else if ctrl == 'q' {
				s.world <- world
				s.turn <- turn
				s.done <- true
				fmt.Println("sent done")
				break turnLoop
			} else if ctrl == 'p' {
				s.world <- world
				s.turn <- turn
			thisloop:
				for {
					keyAgain := <-s.ctrl
					if keyAgain == 'p' {
						s.world <- world
						s.turn <- turn
						break thisloop
					}
				}
			}

		default: // If not, it continues
		}
		for i := startX; i < endX; i++ {
			for j := startY; j < endY; j++ {
				sum := 0
				adj := []int{-1, 0, 1}
				for _, n1 := range adj {
					for _, n2 := range adj {
						if n1 == 0 && n2 == 0 {
						} else if world[(i+n1+p.ImageWidth)%p.ImageWidth][(j+n2+p.ImageHeight)%p.ImageHeight] == 255 {
							sum++
						}
					}
				}
				if (world[i][j] == 255) && (sum < 2 || sum > 3) {
					nextWorld[i-startX][j-startY] = 0
				} else if (world[i][j] == 0) && (sum == 3) {
					nextWorld[i-startX][j-startY] = 255
				} else {
					nextWorld[i-startX][j-startY] = world[i][j]
				}
			}
		}
		// s.world = append([][]byte{}, nextWorld...)
		world = append([][]byte{}, nextWorld...)
		nextWorld = make([][]byte, width)
		for i := range nextWorld {
			nextWorld[i] = make([]byte, height)
		}

	}
	fmt.Println("bottom of generate thing")
	res.WorldPart = world
	res.Turn = turn
	// s.done = true
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
	// done := false

	err := rpc.Register(&GameOfLifeWorker{tick, world, turn, ctrl, done})
	// There is a problem where if we run distributor twice channels are not free the second time because there is something in it
	if err != nil {
		fmt.Println("Error registering listener", err)
		return
	}
	fmt.Println("HERE")
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer func(listener net.Listener) {
		fmt.Println("HERE2")
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener", err)
			return
		} else {
			fmt.Println("closed listener")
		}
	}(listener)
	fmt.Println("HERE3")
	rpc.Accept(listener)
	fmt.Println("HERE4")

}
