package stubs

var GenerateGameOfLife = "Broker.GenerateGameOfLife"
var AliveCellCount = "Broker.AliveCellCountTick"
var Control = "Broker.Control"
var GeneratePart = "Worker.GeneratePart"
var Close = "Worker.Close"
var HaloExchange = "Worker.HaloExchange"
var WorkerAlive = "Worker.Ping"

type BrokerResponse struct {
	World [][]byte
	Turn  int
}

type WorkerResponse struct {
	WorldPart [][]byte
	Complete  bool
}

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type WorldState struct {
	Turn  int
	World [][]byte
}

type Request struct {
	World    [][]byte
	Params   Params
	StartX   int
	EndX     int
	StartY   int
	EndY     int
	ClientID string
}

type AliveResponse struct {
	Alive bool
}

type ControlRequest struct {
	Ctrl rune
}

type HaloRequest struct {
	Halo []byte
}

type HaloResponse struct {
	Halo []byte
}
