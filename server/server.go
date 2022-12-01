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
var globeSlice [][]byte
var turn int
var listener net.Listener
var begun = 0

func (g *GolLogicOperations) CalculateGOL(req stubs.GOLRequest, res *stubs.GOLResponse) (err error) {
	if begun == 1 { //Will skip on first turn - no halos to update.
		updateOnHalos(req.World, req.Start, req.End, req.P) //req.World = halos
	}
	if req.P.Turns == 1 && req.P.ImageHeight == 16 {
		fmt.Println("YEEEOW")
	}
	begun = 1
	//fmt.Println("start: ", req.Start, " end : ", req.End)
	topBottom := calculateNextState(req.P, globe, req.Start, req.End) //Globe was set in setupGol.
	res.Result = topBottom
	res.Turn = turn
	return
}

func updateOnHalos(halos [][]byte, start int, end int, p stubs.Params) {
	//Updating globe based on halos.
	dy := p.ImageHeight / p.Threads
	workerId := start / dy //w0, w1, w2 etc.
	//Adjust the workers halo.
	for j := 0; j < p.ImageWidth; j++ {
		if workerId == 0 && p.Threads == 1 {
			globe[p.ImageHeight-1][j] = halos[len(halos)-1][j] //Bottom row of split above.
			globe[0][j] = halos[0][j]                          //Top of split below.
		} else if workerId == 0 {
			globe[p.ImageHeight-1][j] = halos[len(halos)-1][j] //Bottom row of split above.
			globe[end][j] = halos[2][j]                        //Top of split below.
		} else if workerId == p.Threads-1 {
			globe[start-1][j] = halos[((workerId-1)*2)+1][j]
			globe[0][j] = halos[0][j]
		} else {
			globe[start-1][j] = halos[((workerId-1)*2)+1][j] //Bottom row of split above.
			globe[end][j] = halos[(workerId+1)*2][j]         //Top of split below.

		}
	}
}

func (g *GolLogicOperations) CurrentState(req stubs.NilRequest, res *stubs.StateResponse) (err error) {
	res.World = globe
	res.Turn = turn
	res.WorldSplit = globeSlice
	return
}

func (g *GolLogicOperations) CloseServer(req stubs.NilRequest, res *stubs.NilRequest) (err error) {
	listener.Close()
	return
}

func (g *GolLogicOperations) SetupGol(req stubs.StateResponse, res *stubs.NilRequest) (err error) {
	fmt.Println("resetting globe")
	globe = req.World
	return
}

func calculateNextState(p stubs.Params, world [][]byte, start int, end int) [][]byte {
	newWorld := createBlankState(p)
	result := make([][]byte, end-start)
	for i := 0; i < end-start; i++ {
		result[i] = make([]byte, p.ImageWidth)
	}
	//For storing top and bottom row of slice, used in halo exchange.
	topBottom := make([][]byte, 2)
	topBottom[0] = make([]byte, p.ImageWidth)
	topBottom[1] = make([]byte, p.ImageWidth)
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
			if i == start {
				topBottom[0][j] = newWorld[i][j]
			} else if i == end-1 {
				topBottom[1][j] = newWorld[i][j]
			}

		}
	}
	if p.Turns == 1 && p.ImageHeight == 16 {
		fmt.Println(newWorld[8][4])
		fmt.Println(topBottom)
	}
	globe = newWorld
	globeSlice = result
	return topBottom
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
	ip := flag.String("ip", "127.0.0.1", "ip to listen to")
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&GolLogicOperations{})
	listener, _ = net.Listen("tcp", *ip+":"+*pAddr)
	fmt.Println("connected: ", *ip+":"+*pAddr)
	defer listener.Close()
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
