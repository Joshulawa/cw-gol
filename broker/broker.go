package main

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"uk.ac.bris.cs/gameoflife/stubs"
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
	//Should map workers to an id, deleting workers dynamically means using the
	//index of a worker in the workers list as their id isn't valid.
	//Or is this not a big issue...
	fmt.Println(len(world))
	globe = world
	fmt.Println("hello")
	workers := createWorkers(numbWorkers)
	dy = calculateSplit(numbWorkers, p) //Global variable dy
	turn := 0
	for i := 0; i < p.Turns; i++ { //change back to turn
		turn++
		responses := callGol(numbWorkers, workers, p, world)
		var newWorld [][]byte
		for j := 0; j < numbWorkers; j++ {
			newWorld = append(newWorld, responses[j].Result...) //Think something might be wrong here.
			//OH YEAH NEED TO CHANGE GOL LOGIC IN SERVER TO RETURN ONLY THE RIGHT SIZE SLICE
		}
		world = newWorld
		globe = world
		gturn = turn

		//fmt.Println("wassup dog")
		//Need to include event stuff. ie turn complete.
		//Turn complete stuff would have to be an rpc call to distributor?
	}
	return world
}

func (b *Broker) CurrentState(req stubs.NilRequest, res *stubs.StateResponse) (err error) {
	fmt.Println(len(globe))
	res.World = globe
	res.Turn = gturn
	return
}

func callGol(numbWorkers int, workers []*rpc.Client, p stubs.Params, world [][]byte) []*stubs.GOLResponse {
	responses := make([]*stubs.GOLResponse, numbWorkers)
	results := make([]*rpc.Call, numbWorkers)
	for i := range responses {
		responses[i] = new(stubs.GOLResponse)
	}
	//RPC calls prepared and made in this loop. Responses stored in a list.
	for i, worker := range workers {
		var request stubs.GOLRequest
		if i == len(workers)-1 {
			request = stubs.GOLRequest{p, dy * i, p.ImageHeight, world}
		} else {
			request = stubs.GOLRequest{p, dy * i, dy * (i + 1), world}
		}
		results[i] = worker.Go(stubs.CalculateGOL, request, responses[i], nil)
	}

	//Not pausing long enough for worker.Go to respond?
	for i := 0; i < numbWorkers; i++ {
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
