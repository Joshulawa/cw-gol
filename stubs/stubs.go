package stubs

var CalculateGOL = "GolLogicOperations.CalculateGOL"
var CurrentState = "GolLogicOperations.CurrentState"
var CloseServer = "GolLogicOperations.CloseServer"
var SetupGol = "GolLogicOperations.SetupGol"

var BrokerGol = "Broker.BrokerGol"
var BrokerState = "Broker.CurrentState"

type StartGol struct {
	P       Params
	World   [][]byte
	Workers int
}

type GOLRequest struct {
	P     Params
	Start int
	End   int
	World [][]byte
	//Turn  int
}

type GOLResponse struct {
	Result [][]byte
	Turn   int
}

type StateResponse struct {
	World      [][]byte
	Turn       int
	WorldSplit [][]byte
}

type NilRequest struct{} //Empty request.

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
