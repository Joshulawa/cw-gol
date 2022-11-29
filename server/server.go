package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	//"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type GolLogicOperations struct{}

var globe [][]byte
var turn int
var listener net.Listener

func (g *GolLogicOperations) CalculateGOL(req stubs.GOLRequest, res *stubs.GOLResponse) (err error) {
	world := req.World
	globe = world
	fmt.Println("world before next state : ", len(world))
	world = calculateNextState(req.P, world, req.Start, req.End)
	fmt.Println("world after next state : ", len(world))
	res.Result = world
	res.Turn = turn
	return
}

func (g *GolLogicOperations) CurrentState(req stubs.NilRequest, res *stubs.StateResponse) (err error) {
	res.World = globe
	res.Turn = turn
	return
}

func (g *GolLogicOperations) CloseServer(req stubs.NilRequest, res *stubs.NilRequest) (err error) {
	listener.Close()
	return
}

func calculateNextState(p stubs.Params, world [][]byte, start int, end int) [][]byte {
	newWorld := createBlankState(p)
	result := make([][]byte, end-start)
	for i := 0; i < end-start; i++ {
		result[i] = make([]byte, p.ImageWidth)
	}
	fmt.Println(end)
	for i := start; i < end; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			aliveNeighbours := 0
			//Loop through adjacent cells.
			for a := -1; a <= 1; a++ {
				for b := -1; b <= 1; b++ {
					if world[(p.ImageHeight+i+a)%p.ImageHeight][(p.ImageWidth+j+b)%p.ImageWidth] == 255 {
						if !(a == 0 && b == 0) {
							aliveNeighbours++
						}
					}

				}
			}
			if world[i][j] == 255 && aliveNeighbours < 2 {
				newWorld[i][j] = 0
			} else if world[i][j] == 255 && (aliveNeighbours == 2 || aliveNeighbours == 3) {
				newWorld[i][j] = world[i][j]
			} else if world[i][j] == 255 && aliveNeighbours > 3 {
				newWorld[i][j] = 0
			} else if world[i][j] == 0 && aliveNeighbours == 3 {
				newWorld[i][j] = 255
			} else {
				newWorld[i][j] = world[i][j]
			}
			result[i-start][j] = newWorld[i][j]
		}
	}
	//var newWorld [][]byte
	//for j := 0; j < p.Threads; j++ {
	//	newWorld = append(newWorld, result...)
	//}
	//world = newWorld
	//c.events <- TurnComplete{turn}
	//turn++
	return result
}

func createBlankState(p stubs.Params) [][]byte {
	blankWorld := make([][]byte, p.ImageHeight)
	for i := range blankWorld {
		blankWorld[i] = make([]byte, p.ImageWidth)
	}
	return blankWorld
}

func countAliveCells(p stubs.Params, world [][]byte) int {
	aliveCells := 0
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			if world[i][j] == 0xFF {
				aliveCells++
			}
		}
	}
	return aliveCells
}

func main() {
	//pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&GolLogicOperations{})
	listener, _ = net.Listen("tcp", "18.212.255.211:8030") //"127.0.0.1:"+*pAddr)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)
	rpc.Accept(listener)
}

//for i := start; i < end; i++ {
//for j := 0; j < p.ImageWidth; j++ {
//aliveNeighbours := 0
////Loop through adjacent cells.
//for a := -1; a <= 1; a++ {
//for b := -1; b <= 1; b++ {
//if world[(p.ImageHeight+i+a)%p.ImageHeight][(p.ImageWidth+j+b)%p.ImageWidth] == 255 {
//if !(a == 0 && b == 0) {
//aliveNeighbours++
//}
//}
//
//}
//}
//if world[i][j] == 255 && aliveNeighbours < 2 {
//newWorld[i][j] = 0
//} else if world[i][j] == 255 && (aliveNeighbours == 2 || aliveNeighbours == 3) {
//newWorld[i][j] = world[i][j]
//} else if world[i][j] == 255 && aliveNeighbours > 3 {
//newWorld[i][j] = 0
//} else if world[i][j] == 0 && aliveNeighbours == 3 {
//newWorld[i][j] = 255
//} else {
//newWorld[i][j] = world[i][j]
//}
//
//result[i-start][j] = newWorld[i][j]
//}
