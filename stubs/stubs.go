package stubs

var CalculateGOL = "GolLogicOperations.CalculateGOL"
var CurrentState = "GolLogicOperations.CurrentState"
var CloseServer = "GolLogicOperations.CloseServer"

type GOLResponse struct {
	Result [][]byte
	Turn   int
}

type GOLRequest struct {
	P     Params
	Start int
	End   int
	World [][]byte
	//Turn  int
}

type StateResponse struct {
	World [][]byte
	Turn  int
}

type NilRequest struct{} //Empty request.

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
