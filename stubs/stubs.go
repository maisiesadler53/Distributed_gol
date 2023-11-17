package stubs

var GenerateGameOfLife = "Broker.GenerateGameOfLife"
var AliveCellCount = "Broker.AliveCellCountTick"
var Control = "Broker.Control"
var GeneratePart = "Worker.GeneratePart"
var Close = "Worker.Close"

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
