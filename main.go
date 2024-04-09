package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

func (b* Board) renderBoard() {
	sep := "  "
	for range SIZE {
		sep += "----"
	}
	sep += "-"
	num := " "
	for i := range SIZE {
		num += "   " + fmt.Sprint(i)
	}
	fmt.Printf("%v\n", num)
	fmt.Printf("%v \n", sep)
	for y := range SIZE {
		row := fmt.Sprint(y) + " | "
		for x := range SIZE {
			if b[x][y] == -1 {
				row += color.GreenString("X")
			} else if b[x][y] == 1 {
				row += color.YellowString("O")
			} else {
				row += " "
			}

			if x == SIZE-1 {
				row += " |"
			} else {
				row += " | "
			}
		}
		fmt.Printf("%v \n", row)
		fmt.Printf("%v \n", sep)

	}
}

func (b *Board) validMove(x, y int) bool {
	if x < 0 || x >= SIZE || y < 0 || y >= SIZE {
		return false
	}
	return b[x][y] == 0
}

func (b *Board) checkForWinner(x, y int) int {
	dx := 0
	dy := 0
	for k := 0; k < 4; k++ {
		switch k {
		case 0:
			dx = 1
			dy = 0
		case 1:
			dx = 1
			dy = 1
		case 2:
			dx = 0
			dy = 1
		case 3:
			dx = -1
			dy = 1
		}
		x_ := x
		y_ := y
		j := 0
		streak := 0
		symbol := 0
		for j > -LINE && x_ < SIZE && x_ >= 0 && y_ < SIZE && y_ >= 0 {
			x_ -= dx
			y_ -= dy
			j--
		}
		x_ += dx
		y_ += dy
		j++
		for j < LINE && x_ < SIZE && x_ >= 0 && y_ < SIZE && y_ >= 0 {
			if b[x_][y_] == symbol {
				streak++
			} else {
				streak = 1
				symbol = b[x_][y_]
			}
			if streak == LINE && symbol != 0 {
				return symbol
			}
			x_ += dx
			y_ += dy
			j++
		}

	}
	return 0
}

func (b *Board) goodMoveFor(x, y, player, depth int, stopc <-chan bool) (int, int, int) {
	if depth == 0 { return 0, 0, 1 }
  newBoard := *b
	newBoard[x][y] = player
	winner := newBoard.checkForWinner(x, y)
	if winner != 0 { return winner, 1, 1 }

	komplexityNumber := 0
	bestMove := player
	steps := 0
	for y_ := range SIZE {
		for x_ := range SIZE {
      select {
      case <-stopc:
        goto outer
      default:
      }
			if newBoard.validMove(x_, y_) {
				gmf := 0
				kn := 0
				gmf, kn, steps = newBoard.goodMoveFor(x_, y_, -player, depth-1, stopc)
				komplexityNumber += kn
				if -player == 1 {
					if gmf == 1 {
						return 1, komplexityNumber, steps + 1
					}
					bestMove = max(bestMove, gmf)
				} else {
					if gmf == -1 {
						return -1, komplexityNumber, steps + 1
					}
					bestMove = min(bestMove, gmf)
				}
			}
		}
	}
  outer:
	return bestMove, komplexityNumber, steps + 1
}

func (b *Board) printWinner(x, y int) bool {
	winner := b.checkForWinner(x, y)
	if winner != 0 {
		if winner == 1 {
			fmt.Printf("You win! \n")
		}
		if winner == -1 {
			fmt.Printf("The computer is too smart for you \n")
		}
		return true
	}
	return false
}

func (b *Board) playerTurn() bool {
	fmt.Printf("Players turn: ")
	moveX := 0
	moveY := 0

	_, _ = fmt.Scanf("%d %d \n", &moveX, &moveY)

again:
	if b.validMove(moveX, moveY) {
		b[moveX][moveY] = 1
	} else {
		fmt.Printf("Invalid move\n")
		goto again
	}

	b.renderBoard()

	return b.printWinner(moveX, moveY)
}

type ThreadRes struct {
  x, y, gmf, kn, steps int
}

func (b *Board) computerTurn() bool {
	goodMoveX := -1
	goodMoveY := -1
	neutralMoveX := -1
	neutralMoveY := -1
	notGoodMoveX := -1
	notGoodMoveY := -1

	progress := 0
	knmax := -1
	stepsMin := DEPTH + 1
  c := make(chan ThreadRes)
  stop := make(chan bool)
  start := time.Now()

	gorutinesRunning := 0
	for y := range SIZE {
		for x := range SIZE {
			if b.validMove(x, y) {
        go func(d chan<- ThreadRes){
          gmf, kn, steps := b.goodMoveFor(x, y, -1, DEPTH, stop)
          d <- ThreadRes{ x, y, gmf, kn, steps }
        }(c)
				gorutinesRunning++
			}
			progress++
			fmt.Printf("progress: %v/%v\r", progress, SIZE*SIZE)
		}
	}

	fmt.Printf("collecting results...\n")
	for i := 0; i < gorutinesRunning; i++ {
    res := <- c
    x, y, gmf, kn, steps := res.x, res.y, res.gmf, res.kn, res.steps
		if gmf == -1 && steps < stepsMin {
			fmt.Printf("Computer si playing %v %v\n", x, y)
			goodMoveX = x
			goodMoveY = y
			stepsMin = steps
			fmt.Printf("winmove found on %v %v\n", x, y)
      break
		} else if gmf == 0 {
			if kn > knmax {
				knmax = kn
				neutralMoveX = x
				neutralMoveY = y
			}
		} else if gmf == 1 {
			notGoodMoveX = x
			notGoodMoveY = y
		}
	}
  
  if time.Since(start).Seconds() < 5 {
    DEPTH++
  }

	cmoveX := -1
	cmoveY := -1
	if goodMoveX == -1 {
		if neutralMoveX == -1 {
			if notGoodMoveX == -1 {
				fmt.Printf("Draw!\n")
				return true
			} else {
				cmoveX = notGoodMoveX
				cmoveY = notGoodMoveY
				fmt.Printf("I lost but whatever\n")
			}
		} else {
			cmoveX = neutralMoveX
			cmoveY = neutralMoveY
		}
	} else {
		cmoveX = goodMoveX
		cmoveY = goodMoveY
	}

	fmt.Printf("Computer is plying %v %v\n", cmoveX, cmoveY)
	b[cmoveX][cmoveY] = -1
	b.renderBoard()
	return b.printWinner(cmoveX, cmoveY)
}

type Board [SIZE][SIZE]int

const (
  SIZE = 6
  LINE = 4
)
var (
  DEPTH = 5
)

func main() {

  board := Board{}

	gameEnd := false
	board[SIZE/2][SIZE/2] = -1
	board.renderBoard()
	for {
		gameEnd = board.playerTurn()
		if gameEnd { return }
		gameEnd = board.computerTurn()
		if gameEnd { return }
	}
}
