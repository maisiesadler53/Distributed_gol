package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"testing"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

type args struct {
	p             gol.Params
	expectedAlive []util.Cell
}

type test struct {
	name string
	args args
}

func TestGol(t *testing.T) {
	tests := []test{
		{"0 turns", args{
			p: gol.Params{
				Turns:       0,
				ImageWidth:  16,
				ImageHeight: 16,
			},
			expectedAlive: []util.Cell{
				{X: 4, Y: 5},
				{X: 5, Y: 6},
				{X: 3, Y: 7},
				{X: 4, Y: 7},
				{X: 5, Y: 7},
			},
		}},

		{"1 turn", args{
			p: gol.Params{
				Turns:       1,
				ImageWidth:  16,
				ImageHeight: 16,
			},
			expectedAlive: []util.Cell{
				{X: 3, Y: 6},
				{X: 5, Y: 6},
				{X: 4, Y: 7},
				{X: 5, Y: 7},
				{X: 4, Y: 8},
			},
		}},

		{"100 turns", args{
			p: gol.Params{
				Turns:       100,
				ImageWidth:  16,
				ImageHeight: 16,
			},
			expectedAlive: []util.Cell{
				{X: 12, Y: 0},
				{X: 13, Y: 0},
				{X: 14, Y: 0},
				{X: 13, Y: 14},
				{X: 14, Y: 15},
			},
		}},
	}

	// Run normal tests
	for _, test := range tests {
		for threads := 1; threads <= 16; threads += 1 {
			testName := fmt.Sprintf("%dx%dx%d-%d", test.args.p.ImageWidth, test.args.p.ImageHeight, test.args.p.Turns, threads)
			t.Run(testName, func(t *testing.T) {
				test.args.p.Threads = threads
				aliveCells := make(chan []util.Cell)
				gol.Run(test.args.p, aliveCells, nil)
				var cells []util.Cell
				for newCells := range aliveCells {
					cells = newCells
				}
				assertEqualBoard(t, cells, test.args.expectedAlive, test.args.p)
			})
		}
	}
}

// TestTrace is a special test to be used to generate traces - not a real test
func TestTrace(t *testing.T) {
	traceParams := gol.Params{
		Turns:       10,
		Threads:     4,
		ImageWidth:  64,
		ImageHeight: 64,
	}
	f, _ := os.Create("trace.out")
	aliveCells := make(chan []util.Cell)
	err := trace.Start(f)
	util.Check(err)
	gol.Run(traceParams, aliveCells, nil)
	for range aliveCells {
	}
	trace.Stop()
	err = f.Close()
	util.Check(err)
}

func boardFail(t *testing.T, given, expected []util.Cell, p gol.Params) bool {
	errorString := fmt.Sprintf("-----------------\n\n  FAILED TEST\n  16x16\n  %d Workers\n  %d Turns\n", p.Threads, p.Turns)
	errorString = errorString + util.AliveCellsToString(given, expected, p.ImageWidth, p.ImageHeight)
	t.Error(errorString)
	return false
}

func assertEqualBoard(t *testing.T, given, expected []util.Cell, p gol.Params) bool {
	givenLen := len(given)
	expectedLen := len(expected)

	if givenLen != expectedLen {
		return boardFail(t, given, expected, p)
	}

	visited := make([]bool, expectedLen)
	for i := 0; i < givenLen; i++ {
		element := given[i]
		found := false
		for j := 0; j < expectedLen; j++ {
			if visited[j] {
				continue
			}
			if expected[j] == element {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return boardFail(t, given, expected, p)
		}
	}

	return true
}
