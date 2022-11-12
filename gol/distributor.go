package gol

import (
	"strconv"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

var turn = 0

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	c.ioFilename <- strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth)
	c.ioCommand <- ioInput //This should tell io.go to run readpgm()
	world := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		world[i] = make([]byte, p.ImageWidth)
	}
	//Creating and filling the empty world with data
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			world[row][col] = <-c.ioInput
		}
	}

	//fmt.Println(world)

	//turn := 0

	dy := p.ImageHeight / p.Threads

	var channels [16]chan [][]uint8 //CHANGE TO BE  A MAKE THING FOR CHANNELS AND THEN FOR LOOP ADD THE NUMBER YOU NEED
	//channels := make([]chan [][]byte, 16) //What is the max number of threads?

	for i := 0; i < p.Threads; i++ {
		channels[i] = make(chan [][]uint8)
	}

	// TODO: Execute all turns of the Game of Life.
	for i := 0; i < p.Turns; i++ {
		for i := 0; i < p.Threads; i++ {
			go calculateNextState(dy*i, dy*(i+1), p, world, c, turn)
		}

		for i := 0; i < p.Threads; i++ {
			world = append(world, <-channels[i]...)
		}
		//world = calculateNextState(p, world, c, turn)
		c.events <- TurnComplete{turn}
		turn++
	}
	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          calculateAliveCells(p, world),
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

//func tickTock(p Params, world [][]byte, c distributorChannels) {
//	time.Sleep(2 * time.Second)
//	alive := 0
//	for i := 0; i < p.ImageHeight; i++ {
//		for j := 0; j < p.ImageWidth; j++ {
//			if world[i][j] == 0xFF {
//				alive++
//			}
//		}
//	}
//	c.events <- AliveCellsCount{
//		CompletedTurns: turn,
//		CellsCount:     alive,
//	}
//}

func calculateNextState(start int, end int, p Params, world [][]byte, c distributorChannels, turn int) [][]byte {
	newWorld := createBlankState(p)
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

			if newWorld[i][j] != world[i][j] {
				c.events <- CellFlipped{
					CompletedTurns: turn,
					Cell:           util.Cell{X: j, Y: i},
				}
			}
		}
	}
	return newWorld
}

func createBlankState(p Params) [][]byte {
	blankWorld := make([][]byte, p.ImageHeight)
	for i := range blankWorld {
		blankWorld[i] = make([]byte, p.ImageWidth)
	}
	return blankWorld
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
