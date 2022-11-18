package main

import (
	"flag"
	"net"
	"net/rpc"
	//"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/stubs"
)

//type Params struct {
//	Turns       int
//	Threads     int
//	ImageWidth  int
//	ImageHeight int
//}

func createBlankState(p stubs.Params) [][]byte {
	blankWorld := make([][]byte, p.ImageHeight)
	for i := range blankWorld {
		blankWorld[i] = make([]byte, p.ImageWidth)
	}
	return blankWorld
}

type GolLogicOperations struct{}

func (g *GolLogicOperations) CalculateNextState(req stubs.Request, res *stubs.Response) (err error) {
	p := req.P
	end := req.End
	start := req.Start
	world := req.World

	newWorld := createBlankState(req.P) //Giving this blank world as a param would make the function more efficient.
	result := make([][]byte, end-start)
	for i := 0; i < end-start; i++ {
		result[i] = make([]byte, p.ImageWidth)
	}
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
	res.Result = result
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&GolLogicOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
