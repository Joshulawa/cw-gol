package stubs

var CalculateGOL = "GolLogicOperations.CalculateGOL"
var CalculateAliveCells = "GolLogicOperations.CalculateAliveCells"

type GOLResponse struct {
	Result [][]byte
}

type GOLRequest struct {
	P     Params
	Start int
	End   int
	World [][]byte
	//Turn  int
}

type AliveCellResponse struct {
	Alive int
	Turn  int
}

type NilRequest struct{} //Empty request.

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
