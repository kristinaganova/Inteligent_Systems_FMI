package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

type Move int

const (
	Left Move = iota
	Right
	Up
	Down
)

type Puzzle struct {
	size  int
	tiles []int
	zero  int
}

type GoalInfo struct {
	target    []int
	positions map[int]int
	zero      int
}

func readInput() (int, int, []int, error) {
	in := bufio.NewReader(os.Stdin)
	var N int
	if _, err := fmt.Fscan(in, &N); err != nil {
		return 0, 0, nil, err
	}

	total := N + 1
	k := int(math.Round(math.Sqrt(float64(total))))
	if k*k != total {
		return 0, 0, nil, fmt.Errorf("invalid N: not a k*k-1 puzzle")
	}

	var I int
	if _, err := fmt.Fscan(in, &I); err != nil {
		return 0, 0, nil, err
	}

	tiles := make([]int, total)
	for i := 0; i < total; i++ {
		if _, err := fmt.Fscan(in, &tiles[i]); err != nil {
			return 0, 0, nil, err
		}
	}

	return N, I, tiles, nil
}

func parsePuzzle(tiles []int, N int, _ int) (Puzzle, error) {
	total := N + 1
	k := int(math.Round(math.Sqrt(float64(total))))

	zero := -1
	for i, v := range tiles {
		if v == 0 {
			zero = i
			break
		}
	}
	if zero == -1 {
		return Puzzle{}, fmt.Errorf("missing 0 in the arrangement")
	}
	return Puzzle{size: k, tiles: tiles, zero: zero}, nil
}

func parseGoal(_ []int, N int, I int) (GoalInfo, error) {
	total := N + 1

	zeroIdx := total - 1
	if I >= 0 && I < total {
		zeroIdx = I
	}

	goal := make([]int, total)
	used := make([]bool, total)

	goal[zeroIdx] = 0
	used[zeroIdx] = true

	num := 1
	for i := 0; i < total; i++ {
		if used[i] {
			continue
		}
		goal[i] = num
		num++
	}

	positions := make(map[int]int, total)
	for i, v := range goal {
		positions[v] = i
	}

	return GoalInfo{target: goal, positions: positions, zero: zeroIdx}, nil
}

func countInversions(tiles []int) int {
	arr := make([]int, 0, len(tiles)-1)
	for _, v := range tiles {
		if v != 0 {
			arr = append(arr, v)
		}
	}
	inversions := 0
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] > arr[j] {
				inversions++
			}
		}
	}
	return inversions
}

func getBlankRowFromBottom(_ []int, size int, zeroIdx int) int {
	return size - (zeroIdx / size)
}

func isSolvable(puzzle Puzzle, goal GoalInfo) bool {
	k := puzzle.size

	invStart := countInversions(puzzle.tiles)
	rowFromBottomStart := getBlankRowFromBottom(puzzle.tiles, k, puzzle.zero)

	invGoal := countInversions(goal.target)
	rowFromBottomGoal := getBlankRowFromBottom(goal.target, k, goal.zero)

	if k%2 == 0 {
		return ((invStart + rowFromBottomStart) % 2) == ((invGoal + rowFromBottomGoal) % 2)
	}
	return (invStart % 2) == (invGoal % 2)
}

func indexToCoords(index int, size int) (int, int) {
	return index / size, index % size
}

func coordsToIndex(row int, col int, size int) int {
	return row*size + col
}

func iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func manhattanDistance(start Puzzle, goal GoalInfo) int {
	distance := 0
	for i, v := range start.tiles {
		if v == 0 {
			continue
		}
		targetIdx := goal.positions[v]
		tr, tc := indexToCoords(targetIdx, start.size)
		r, c := indexToCoords(i, start.size)
		distance += iabs(r-tr) + iabs(c-tc)
	}
	return distance
}

func neighbors(start Puzzle, lastMove Move) []Move {
	moves := make([]Move, 0, 4)
	row, col := indexToCoords(start.zero, start.size)
	if col > 0 && lastMove != Right {
		moves = append(moves, Left)
	}
	if col < start.size-1 && lastMove != Left {
		moves = append(moves, Right)
	}
	if row > 0 && lastMove != Down {
		moves = append(moves, Up)
	}
	if row < start.size-1 && lastMove != Up {
		moves = append(moves, Down)
	}
	return moves
}

func doMove(start Puzzle, move Move) Puzzle {
	k := start.size
	row, col := indexToCoords(start.zero, k)

	switch move {
	case Left:
		col--
	case Right:
		col++
	case Up:
		row--
	case Down:
		row++
	}

	newIndex := coordsToIndex(row, col, k)
	newTiles := make([]int, len(start.tiles))
	copy(newTiles, start.tiles)
	newTiles[start.zero], newTiles[newIndex] = newTiles[newIndex], newTiles[start.zero]

	return Puzzle{size: k, tiles: newTiles, zero: newIndex}
}

func dfs(start Puzzle, goal GoalInfo, gCost int, limit int, lastMove Move, path *[]Move) (bool, int) {
	h := manhattanDistance(start, goal)
	f := gCost + h

	if f > limit {
		return false, f
	}
	if h == 0 {
		return true, gCost
	}

	minExcess := math.MaxInt
	for _, move := range neighbors(start, lastMove) {
		next := doMove(start, move)
		*path = append(*path, move)
		found, nextLimit := dfs(next, goal, gCost+1, limit, move, path)
		if found {
			return true, nextLimit
		}

		minExcess = min(minExcess, nextLimit)
		*path = (*path)[:len(*path)-1]
	}
	return false, minExcess
}

func solve(start Puzzle, goal GoalInfo) ([]Move, bool) {
	limit := manhattanDistance(start, goal)
	path := make([]Move, 0, 256)

	for {
		found, nextLimit := dfs(start, goal, 0, limit, -1, &path)
		if found {
			return path, true
		}
		if nextLimit == math.MaxInt {
			return nil, false
		}
		limit = nextLimit
	}
}

func main() {
	N, I, tiles, err := readInput()
	if err != nil {
		fmt.Println("Failed to read input:", err)
		return
	}

	puzzle, err := parsePuzzle(tiles, N, I)
	if err != nil {
		fmt.Println("Failed to parse puzzle:", err)
		return
	}

	goal, err := parseGoal(tiles, N, I)
	if err != nil {
		fmt.Println("Failed to parse goal:", err)
		return
	}

	if !isSolvable(puzzle, goal) {
		fmt.Println(-1)
		return
	}

	path, found := solve(puzzle, goal)
	if !found {
		fmt.Println(-1)
		return
	}

	fmt.Println(len(path))
	moveNames := map[Move]string{
		Left:  "right",
		Right: "left",
		Up:    "down",
		Down:  "up",
	}
	for _, m := range path {
		fmt.Println(moveNames[m])
	}
}
