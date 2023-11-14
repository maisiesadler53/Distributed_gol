package stubs

var GenerateGameOfLife = "GameOfLifeWorker.GenerateGameOfLife"
var AliveCellCount = "GameOfLifeWorker.AliveCellCountTick"
var Control = "GameOfLifeWorker.Control"

type Response struct {
	WorldPart [][]byte
	Turn      int
}
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Request struct {
	World  [][]byte
	Params Params
	StartX int
	EndX   int
	StartY int
	EndY   int
	Ctrl   rune
}
