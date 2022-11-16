package gol

import (
	"fmt"
	"strconv"
	"time"

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
			calcCellFlipped(c, p, world, blankWorld, 0, row, col) //Shouldnt call this each time really
		}
	}

	dy := p.ImageHeight / p.Threads

	//List for channels used by workers.
	var channels [32]chan [][]uint8 //CHANGE TO BE  A MAKE THING FOR CHANNELS AND THEN FOR LOOP ADD THE NUMBER YOU NEED
	//What is the max number of threads?

	for i := 0; i < p.Threads; i++ {
		channels[i] = make(chan [][]uint8)
	}

	turn := 0

	// TODO: Execute all turns of the Game of Life.
	worldChannel := make(chan [][]byte)
	turnChannel := make(chan int)
	worldChannel2 := make(chan [][]byte)
	turnChannel2 := make(chan int)
	quit := make(chan bool)
	pause := make(chan bool)

	go tickTock(worldChannel, turnChannel, p, c) //Start the ticker.
	go keyInput(p, c, worldChannel2, turnChannel2, quit, pause)

	for i := 0; i < p.Turns; i++ {
		select {
		case <-quit:
			fmt.Println("wassup")
			i = p.Turns //This
		case <-pause:
			<-pause //This should block until pause is pressed again
		default:
			//Non blocking send.
			//When should I send world down the worldChannel , beggining or end?
			select {
			case worldChannel <- world:
				worldChannel2 <- world
				turnChannel <- turn
				turnChannel2 <- turn
			default:
			}
			for j := 0; j < p.Threads; j++ {
				if j == p.Threads-1 {
					go calculateNextState(dy*j, p.ImageHeight, p, world, c, turn, channels[j])
				} else {
					go calculateNextState(dy*j, dy*(j+1), p, world, c, turn, channels[j])
				}
			}
			var newWorld [][]byte
			for j := 0; j < p.Threads; j++ {

				newWorld = append(newWorld, <-channels[j]...)
			}
			world = newWorld
			c.events <- TurnComplete{turn}
			turn++
		}

	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          calculateAliveCells(p, world),
	}

	imageOutput(p, c, world, turn)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func imageOutput(p Params, c distributorChannels, world [][]byte, turn int) {

	c.ioFilename <- strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(turn)
	c.ioCommand <- ioOutput

	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- world[row][col]
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
	//postWorld(newWorld)
	//fmt.Println("yo  ", result)
	channel <- result //NEED TO ONLY RETURN THE RIGHT PARTS
}

func createBlankState(p Params) [][]byte {
	blankWorld := make([][]byte, p.ImageHeight)
	for i := range blankWorld {
		blankWorld[i] = make([]byte, p.ImageWidth)
	}
	return blankWorld
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

func tickTock(worldChannel chan [][]byte, turnChannel chan int, p Params, c distributorChannels) {
	for {
		time.Sleep(2 * time.Second)
		alive := 0
		world := <-worldChannel
		turn := <-turnChannel
		for i := 0; i < p.ImageHeight; i++ {
			for j := 0; j < p.ImageWidth; j++ {
				if world[i][j] == 0xFF {
					alive++
				}
			}
		}
		c.events <- AliveCellsCount{
			CompletedTurns: turn,
			CellsCount:     alive,
		}
	}

}
