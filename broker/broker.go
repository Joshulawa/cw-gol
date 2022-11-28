package main

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var gturn int
var globe [][]byte
var dy int //Global variable so it can be changed during iterations of game of life?
//Think of fault tolerance...

//Fault tolerance -- If a worker quits mid-turn, need to cancel/invalidate that turn
//and do it again with the reformatted world distribution.
//Should try and interpret flags.

type Broker struct{}

func (b *Broker) BrokerGol(req stubs.StartGol, res *stubs.GOLResponse) (err error) {
	fmt.Println("Distributor connected")
	res.Result = distribute(req.P, req.World, req.P.Threads)
	res.Turn = gturn
	//res.Result = req.World
	//Need to return number of turns completed at some point as well.
	return
}

func distribute(p stubs.Params, world [][]byte, numbWorkers int) [][]byte {
	globe = world
	fmt.Println("inside distribute")
	workers := createWorkers(numbWorkers) //Creates and connects clients to servers.
	//Pass world to all servers to setupGol.
	for _, worker := range workers {
		request := stubs.StateResponse{
			World: world,
			Turn:  0,
		}
		worker.Call(stubs.SetupGol, request, new(stubs.NilRequest))
	}

	dy = calculateSplit(numbWorkers, p) //Global variable dy
	turn := 0
	if p.ImageWidth == 16 {
		util.VisualiseMatrix(world, p.ImageWidth, p.ImageHeight)
	}

	responses := make([]*stubs.GOLResponse, numbWorkers)
	halos := make([][]byte, numbWorkers*2)
	for i := 0; i < numbWorkers*2; i++ {
		halos[i] = make([]byte, p.ImageWidth)
	}
	for a := 0; a < p.Turns; a++ {

		turn++
		//send halos instead of world
		responses = haloGol(workers, p, halos)
		var newHalos [][]byte
		for j := range responses {
			newHalos = append(newHalos, responses[j].Result...)

		}
		if p.Turns == 1 && p.ImageHeight == 16 {
			fmt.Println(newHalos)
		}
		halos = newHalos
	}
	if p.Turns == 0 {
		return globe
	}
	world = getResults(workers)
	return world
}

func getResults(workers []*rpc.Client) [][]byte {
	var newWorld [][]byte
	for _, worker := range workers {
		response := new(stubs.StateResponse)
		worker.Call(stubs.CurrentState, stubs.NilRequest{}, response)
		newWorld = append(newWorld, response.WorldSplit...)
	}
	return newWorld
}

func (b *Broker) CurrentState(req stubs.NilRequest, res *stubs.StateResponse) (err error) {
	res.World = globe
	res.Turn = gturn
	return
}

func haloGol(workers []*rpc.Client, p stubs.Params, halos [][]byte) []*stubs.GOLResponse {
	responses := make([]*stubs.GOLResponse, len(workers))
	results := make([]*rpc.Call, len(workers))
	for i := range responses {
		responses[i] = new(stubs.GOLResponse)
	}
	for i, worker := range workers {
		var request stubs.GOLRequest
		if i == len(workers)-1 {
			request = stubs.GOLRequest{p, dy * i, p.ImageHeight, halos}
		} else {
			request = stubs.GOLRequest{p, dy * i, dy * (i + 1), halos}
		}
		//Each server will return [][]byte of just their halos.
		results[i] = worker.Go(stubs.CalculateGOL, request, responses[i], nil)
	}
	for i := 0; i < len(workers); i++ {
		<-results[i].Done
	}
	return responses
}

func createWorkers(numbWorkers int) []*rpc.Client {
	workers := make([]*rpc.Client, numbWorkers) //Create list of clients.
	for i := range workers {
		workers[i], _ = rpc.Dial("tcp", "127.0.0.1:"+strconv.Itoa(8010+i*10))
		fmt.Println("127.0.0.1:" + strconv.Itoa(8010+i*10))
	}
	return workers
}

//Made as a function so dy can be dynamically alterated and calculated more easily.
func calculateSplit(numbWorkers int, p stubs.Params) int {
	//dy = p.ImageHeight / numbWorkers //where dy is global.
	return p.ImageHeight / numbWorkers
}

func main() {
	fmt.Println("yo")
	err := rpc.Register(&Broker{})
	if err != nil {
		fmt.Println(err)
		return
	}
	listener, _ := net.Listen("tcp", "127.0.0.1:8000")
	defer listener.Close()
	rpc.Accept(listener)
}
