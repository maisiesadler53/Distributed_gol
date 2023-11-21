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

type Worker struct {
	closeListener chan bool
}

func (s *Worker) Close(req stubs.Request, res *stubs.Response) (err error) {
	s.closeListener <- true
	return
}

func (s *Worker) GeneratePart(req stubs.Request, res *stubs.Response) (err error) {
	p := req.Params
	world := append([][]byte{}, req.World...)
	startX := req.StartX
	startY := req.StartY
	endY := req.EndY
	endX := req.EndX
	height := endY - startY
	width := endX - startX
	nextWorldPart := make([][]byte, width)
	for i := range nextWorldPart {
		nextWorldPart[i] = make([]byte, height)
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
				nextWorldPart[i-startX][j-startY] = 0

			} else if (world[i][j] == 0) && (sum == 3) {
				nextWorldPart[i-startX][j-startY] = 255

			} else {
				nextWorldPart[i-startX][j-startY] = world[i][j]
			}
		}
	}
	res.WorldPart = append([][]byte{}, nextWorldPart...)
	return
}

func main() {
	pAddr := flag.String("port", "8000", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	closeListener := make(chan bool, 1)

	//register rpc calls
	err := rpc.Register(&Worker{closeListener: closeListener})
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
	time.Sleep(2 * time.Second)
	return
}