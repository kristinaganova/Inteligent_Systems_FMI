package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	SIZE      = 3
	CELLS     = SIZE * SIZE
	INF       = 1000
	WIN_SCORE = 10
	EMPTY     = '_'
	X         = 'X'
	O         = 'O'
)

type Board [CELLS]rune

func idx(row, col int) int {
	return row*SIZE + col
}

func rc(i int) (int, int) {
	return i / SIZE, i % SIZE
}

func otherPlayer(p rune) rune {
	if p == X {
		return O
	}
	return X
}

func checkWinner(b *Board) rune {
	// rows
	for r := 0; r < SIZE; r++ {
		i0 := idx(r, 0)
		i1 := idx(r, 1)
		i2 := idx(r, 2)
		if b[i0] != EMPTY && b[i0] == b[i1] && b[i1] == b[i2] {
			return b[i0]
		}
	}
	// cols
	for c := 0; c < SIZE; c++ {
		i0 := idx(0, c)
		i1 := idx(1, c)
		i2 := idx(2, c)
		if b[i0] != EMPTY && b[i0] == b[i1] && b[i1] == b[i2] {
			return b[i0]
		}
	}
	// main diag
	if b[idx(0, 0)] != EMPTY &&
		b[idx(0, 0)] == b[idx(1, 1)] &&
		b[idx(1, 1)] == b[idx(2, 2)] {
		return b[idx(0, 0)]
	}
	// second diag
	if b[idx(0, 2)] != EMPTY &&
		b[idx(0, 2)] == b[idx(1, 1)] &&
		b[idx(1, 1)] == b[idx(2, 0)] {
		return b[idx(0, 2)]
	}
	return 0
}

func isBoardFull(b *Board) bool {
	for i := 0; i < CELLS; i++ {
		if b[i] == EMPTY {
			return false
		}
	}
	return true
}

func evaluateTerminal(b *Board, maximizingFor rune, depth int) (bool, int) {
	winner := checkWinner(b)
	if winner != 0 {
		if winner == maximizingFor {
			return true, WIN_SCORE - depth
		}
		return true, depth - WIN_SCORE
	}
	if isBoardFull(b) {
		return true, 0
	}
	return false, 0
}

func maxValue(b *Board, currentTurn, maximizingFor rune, depth, alpha, beta int) int {
	if term, score := evaluateTerminal(b, maximizingFor, depth); term {
		return score
	}

	best := -INF
	for i := 0; i < CELLS; i++ {
		if b[i] == EMPTY {
			b[i] = currentTurn
			score := minValue(b, otherPlayer(currentTurn), maximizingFor, depth+1, alpha, beta)
			b[i] = EMPTY
			if score > best {
				best = score
			}
			if best > alpha {
				alpha = best
			}
			if alpha >= beta {
				break
			}
		}
	}
	return best
}

func minValue(b *Board, currentTurn, maximizingFor rune, depth, alpha, beta int) int {
	if term, score := evaluateTerminal(b, maximizingFor, depth); term {
		return score
	}

	best := INF
	for i := 0; i < CELLS; i++ {
		if b[i] == EMPTY {
			b[i] = currentTurn
			score := maxValue(b, otherPlayer(currentTurn), maximizingFor, depth+1, alpha, beta)
			b[i] = EMPTY
			if score < best {
				best = score
			}
			if best < beta {
				beta = best
			}
			if alpha >= beta {
				break
			}
		}
	}
	return best
}

func minimax(b *Board, currentTurn, maximizingFor rune, depth, alpha, beta int) int {
	if currentTurn == maximizingFor {
		return maxValue(b, currentTurn, maximizingFor, depth, alpha, beta)
	}
	return minValue(b, currentTurn, maximizingFor, depth, alpha, beta)
}

func findBestMove(b *Board, player rune) (int, int) {
	bestScore := -INF
	bestCell := -1

	order := []int{4, 0, 2, 6, 8, 1, 3, 5, 7}

	for _, i := range order {
		if b[i] == EMPTY {
			b[i] = player
			score := minimax(b, otherPlayer(player), player, 1, -INF, INF)
			b[i] = EMPTY
			if score > bestScore {
				bestScore = score
				bestCell = i
			}
		}
	}

	if bestCell == -1 {
		return -1, -1
	}
	r, c := rc(bestCell)
	return r, c
}

func printBoard(b *Board) {
	fmt.Println("+---+---+---+")
	for r := 0; r < SIZE; r++ {
		fmt.Printf("| %c | %c | %c |\n",
			b[idx(r, 0)], b[idx(r, 1)], b[idx(r, 2)],
		)
		fmt.Println("+---+---+---+")
	}
}

func parseBoard(lines []string) Board {
	var b Board
	row := 0
	for lineIdx := 1; lineIdx < 7; lineIdx += 2 {
		line := lines[lineIdx]
		b[idx(row, 0)] = rune(line[2])
		b[idx(row, 1)] = rune(line[6])
		b[idx(row, 2)] = rune(line[10])
		row++
	}
	return b
}

func currentTurnFromBoard(first rune, b *Board) rune {
	moves := 0
	for i := 0; i < CELLS; i++ {
		if b[i] == X || b[i] == O {
			moves++
		}
	}
	if moves%2 == 0 {
		return first
	}
	return otherPlayer(first)
}

func handleJudge(scanner *bufio.Scanner) {
	// TURN X / TURN O
	if !scanner.Scan() {
		return
	}
	line := scanner.Text()
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return
	}
	turn := rune(parts[1][0])

	// 7 lines of board
	lines := make([]string, 0, 7)
	for i := 0; i < 7; i++ {
		if !scanner.Scan() {
			return
		}
		lines = append(lines, scanner.Text())
	}
	board := parseBoard(lines)

	// terminal -> -1
	if w := checkWinner(&board); w != 0 || isBoardFull(&board) {
		fmt.Println(-1)
		return
	}

	row, col := findBestMove(&board, turn)
	fmt.Printf("%d %d\n", row+1, col+1)
}

func handleGame(scanner *bufio.Scanner) {
	// FIRST X / FIRST O
	if !scanner.Scan() {
		return
	}
	lineFirst := scanner.Text()
	partsFirst := strings.Fields(lineFirst)
	if len(partsFirst) < 2 {
		return
	}
	first := rune(partsFirst[1][0])

	// HUMAN X / HUMAN O
	if !scanner.Scan() {
		return
	}
	lineHuman := scanner.Text()
	partsHuman := strings.Fields(lineHuman)
	if len(partsHuman) < 2 {
		return
	}
	human := rune(partsHuman[1][0])
	agent := otherPlayer(human)

	// 7 lines of board
	lines := make([]string, 0, 7)
	for i := 0; i < 7; i++ {
		if !scanner.Scan() {
			return
		}
		lines = append(lines, scanner.Text())
	}
	board := parseBoard(lines)

	currentTurn := currentTurnFromBoard(first, &board)

	for {
		if currentTurn == human {
			if !scanner.Scan() {
				return
			}
			moveLine := scanner.Text()
			fields := strings.Fields(moveLine)
			if len(fields) < 2 {
				continue
			}
			r, err1 := strconv.Atoi(fields[0])
			c, err2 := strconv.Atoi(fields[1])
			if err1 != nil || err2 != nil {
				continue
			}
			r--
			c--
			if r < 0 || r >= SIZE || c < 0 || c >= SIZE {
				continue
			}
			cell := idx(r, c)
			if board[cell] != EMPTY {
				continue
			}
			board[cell] = human
		} else {
			r, c := findBestMove(&board, agent)
			if r >= 0 && c >= 0 {
				board[idx(r, c)] = agent
			}
		}

		printBoard(&board)

		w := checkWinner(&board)
		if w != 0 {
			fmt.Printf("WINNER: %c\n", w)
			return
		}
		if isBoardFull(&board) {
			fmt.Println("DRAW")
			return
		}

		currentTurn = otherPlayer(currentTurn)
	}
}

func main() {
	if os.Getenv("FMI_TIME_ONLY") == "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		return
	}
	mode := strings.TrimSpace(scanner.Text())

	if mode == "JUDGE" {
		handleJudge(scanner)
	} else if mode == "GAME" {
		handleGame(scanner)
	}
}
