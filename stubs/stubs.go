package stubs

var CalculateNextState = "golLogicOperations.CalculateNextState"

type Response struct {
	Result [][]byte
}

type Request struct {
	P     Params
	Start int
	End   int
	World [][]byte
	Turn  int
}

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
