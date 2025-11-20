package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type solver struct {
	n           int
	queens      []int
	rowCnt      []int
	diagMainCnt []int
	diagAntiCnt []int
}

func newSolver(n int) *solver {
	s := &solver{
		n:           n,
		queens:      make([]int, n),
		rowCnt:      make([]int, n),
		diagMainCnt: make([]int, 2*n-1),
		diagAntiCnt: make([]int, 2*n-1),
	}
	for i := range s.queens {
		s.queens[i] = -1
	}
	s.initBoard()
	return s
}

func (s *solver) initBoard() {
	col := 1
	for row := 0; row < s.n; row++ {
		s.queens[col] = row
		s.rowCnt[row]++
		s.diagMainCnt[row-col+s.n-1]++
		s.diagAntiCnt[row+col]++
		col += 2
		if col >= s.n {
			col = 0
		}
	}
}

func (s *solver) place(col, row int) {
	if s.queens[col] != -1 {
		if s.queens[col] == row {
			return
		}
		s.remove(col, s.queens[col])
	}
	s.queens[col] = row
	s.rowCnt[row]++
	s.diagMainCnt[row-col+s.n-1]++
	s.diagAntiCnt[row+col]++
}

func (s *solver) remove(col, row int) {
	s.rowCnt[row]--
	s.diagMainCnt[row-col+s.n-1]--
	s.diagAntiCnt[row+col]--
}

func (s *solver) conflictsAt(row, col int) int {
	c := s.rowCnt[row] + s.diagMainCnt[row-col+s.n-1] + s.diagAntiCnt[row+col]
	if s.queens[col] == row {
		c -= 3
	}
	return c
}

func (s *solver) colWithMaxConflicts() int {
	maxC := -1
	cands := make([]int, 0, s.n)
	for col := 0; col < s.n; col++ {
		row := s.queens[col]
		c := s.conflictsAt(row, col)
		if c > maxC {
			maxC = c
			cands = cands[:0]
			cands = append(cands, col)
		} else if c == maxC {
			cands = append(cands, col)
		}
	}
	return cands[rand.Intn(len(cands))]
}

func (s *solver) rowWithMinConflicts(col int) int {
	minC := int(1<<31 - 1)
	cands := make([]int, 0, s.n)
	for row := 0; row < s.n; row++ {
		c := s.conflictsAt(row, col)
		if c < minC {
			minC = c
			cands = cands[:0]
			cands = append(cands, row)
		} else if c == minC {
			cands = append(cands, row)
		}
	}
	return cands[rand.Intn(len(cands))]
}

func (s *solver) hasConflicts() bool {
	for col := 0; col < s.n; col++ {
		if s.conflictsAt(s.queens[col], col) > 0 {
			return true
		}
	}
	return false
}

func (s *solver) solve() []int {
	if s.n <= 3 {
		return nil
	}
	maxRestarts := 8
	maxSteps := 5 * s.n

	for r := 0; r < maxRestarts; r++ {
		if r > 0 {
			s.resetRandom()
		}
		for step := 0; step < maxSteps; step++ {
			if !s.hasConflicts() {
				return s.queens
			}
			col := s.colWithMaxConflicts()
			row := s.rowWithMinConflicts(col)
			s.place(col, row)
		}
	}
	return nil
}

func (s *solver) resetRandom() {
	for i := range s.rowCnt {
		s.rowCnt[i] = 0
	}
	for i := range s.diagMainCnt {
		s.diagMainCnt[i] = 0
	}
	for i := range s.diagAntiCnt {
		s.diagAntiCnt[i] = 0
	}
	for i := range s.queens {
		s.queens[i] = -1
	}
	cols := make([]int, s.n)
	for i := range cols {
		cols[i] = i
	}
	rand.Shuffle(s.n, func(i, j int) { cols[i], cols[j] = cols[j], cols[i] })
	rows := make([]int, s.n)
	for i := 0; i < s.n; i++ {
		rows[i] = i
	}
	rand.Shuffle(s.n, func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
	for i, col := range cols {
		row := rows[i]
		s.queens[col] = row
		s.rowCnt[row]++
		s.diagMainCnt[row-col+s.n-1]++
		s.diagAntiCnt[row+col]++
	}
}

func printBoard(sol []int) {
	n := len(sol)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if c > 0 {
				fmt.Print(" ")
			}
			if sol[c] == r {
				fmt.Print("*")
			} else {
				fmt.Print("_")
			}
		}
		fmt.Println()
	}
}

func printArray(sol []int) {
	b := strings.Builder{}
	b.Grow(3 * nGuess(len(sol)))
	b.WriteByte('[')
	for i, v := range sol {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(v))
	}
	b.WriteByte(']')
	fmt.Println(b.String())
}

func nGuess(n int) int {
	if n <= 0 {
		return 0
	}
	return n * 4
}

func main() {
	rand.Seed(time.Now().UnixNano())

	boardFlag := flag.Bool("board", false, "print board with '*' and '_'")
	flag.Parse()

	in := bufio.NewReader(os.Stdin)
	var n int
	if _, err := fmt.Fscan(in, &n); err != nil {
		return
	}

	timeOnly := os.Getenv("FMI_TIME_ONLY") == "1"
	start := time.Now()

	if n == 2 || n == 3 {
		elapsed := time.Since(start)
		if timeOnly {
			fmt.Printf("# TIMES_MS: alg=%d\n", elapsed.Milliseconds())
			return
		}
		fmt.Println(-1)
		return
	}
	if n == 1 {
		elapsed := time.Since(start)
		if timeOnly {
			fmt.Printf("# TIMES_MS: alg=%d\n", elapsed.Milliseconds())
			return
		}
		fmt.Println("[0]")
		return
	}

	if timeOnly {
		elapsed := time.Since(start)
		fmt.Printf("# TIMES_MS: alg=%d\n", elapsed.Milliseconds())
		return
	}

	s := newSolver(n)
	sol := s.solve()
	if sol == nil {
		fmt.Println(-1)
		return
	}

	if *boardFlag {
		printBoard(sol)
	} else {
		printArray(sol)
	}
}
