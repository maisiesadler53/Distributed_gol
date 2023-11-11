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
	ticker chan bool
	world  chan [][]byte
}

func (s *GameOfLifeWorker) AliveCellCountTick(req stubs.Request, res *stubs.Response) (err error) {
	// res.WorldPart = s.world
	s.ticker <- true
	res.WorldPart = <-s.world
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

	for turn := 0; turn < p.Turns; turn++ {
		select {
		case <-s.ticker:
			s.world <- world
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
	res.WorldPart = world
	return
}

func main() {
	pAddr := flag.String("port", "8040", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	world := make(chan [][]byte)
	ticker := make(chan bool)

	err := rpc.Register(&GameOfLifeWorker{ticker, world})
	if err != nil {
		fmt.Println("Error registering listener", err)
		return
	}
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener", err)
			return
		}
	}(listener)
	rpc.Accept(listener)
}
