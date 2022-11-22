package gol

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

var server = "127.0.0.1:8030" //flag.String("server", "localhost:8030", "IP:port string to connect to as server")

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	filename := strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth)
	c.ioFilename <- filename
	c.ioCommand <- ioInput

	world := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		world[i] = make([]byte, p.ImageWidth)
	}

	//Creating and filling the empty world with data
	blankWorld := createBlankState(p)
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput
			calcCellFlipped(c, p, world, blankWorld, 0, row, col) //Shouldn't call this each time really
		}
	}

	//dy := p.ImageHeight / p.Threads

	client, _ := rpc.Dial("tcp", "127.0.0.1:8030")
	turn := 0
	request := stubs.GOLRequest{P: stubs.Params(p), Start: 0, End: p.ImageHeight, World: world}
	response := new(stubs.GOLResponse)
	//RPC call to calculate game of life.
	ticker := time.NewTicker(2 * time.Second)
	go tickTock(client, c)
	a := client.Go(stubs.CalculateGOL, request, response, nil)
label:
	for {
		select {
		case <-ticker.C:
			tickTock(client, c)
		case <-a.Done:
			break label
		}
	}

	world = response.Result

	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          calculateAliveCells(p, world),
	}

	imageOutput(p, c, world, p.Turns) //Note the change from turn to p.Turns

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func createBlankState(p Params) [][]byte {
	blankWorld := make([][]byte, p.ImageHeight)
	for i := range blankWorld {
		blankWorld[i] = make([]byte, p.ImageWidth)
	}
	return blankWorld
}

func tickTock(client *rpc.Client, c distributorChannels) {
	request := stubs.NilRequest{}
	response := new(stubs.AliveCellResponse)
	client.Call(stubs.CalculateAliveCells, request, response)
	c.events <- AliveCellsCount{
		CompletedTurns: response.Turn,
		CellsCount:     response.Alive,
	}
}

//-------------------------------------------------------------------------//

func imageOutput(p Params, c distributorChannels, world [][]byte, turn int) {

	c.ioFilename <- strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(turn)
	c.ioCommand <- ioOutput

	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col]
		}
	}
}

func calcCellFlipped(c distributorChannels, p Params, world [][]byte, newWorld [][]byte, turn int, i int, j int) {
	if newWorld[i][j] != world[i][j] {
		c.events <- CellFlipped{
			CompletedTurns: turn,
			Cell:           util.Cell{X: j, Y: i},
		}
	}
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	var alive []util.Cell
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			if world[i][j] == 0xFF {
				alive = append(alive, util.Cell{X: j, Y: i})
			}
		}
	}
	return alive
}

func keyInput(p Params, c distributorChannels, worldChannel chan [][]byte, turnChannel chan int, quit chan bool, pause chan bool) {

	world := <-worldChannel
	turn := <-turnChannel

	for {
		select {
		case world = <-worldChannel:
			turn = <-turnChannel
		case key := <-c.keyPresses:
			if key == 's' {
				fmt.Println("s pressed")
				imageOutput(p, c, world, turn)
			} else if key == 'q' {
				fmt.Println("q pressed")
				quit <- true
			} else if key == 'p' {
				pause <- true
			}
		}

	}
}

func calculateNextState(start int, end int, p Params, world [][]byte, c distributorChannels, turn int, channel chan [][]byte) {
	newWorld := createBlankState(p) //Giving this blank world as a param would make the function more efficient.
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

			calcCellFlipped(c, p, world, newWorld, turn, i, j)

			result[i-start][j] = newWorld[i][j]
		}
	}
	channel <- result //NEED TO ONLY RETURN THE RIGHT PARTS
}
