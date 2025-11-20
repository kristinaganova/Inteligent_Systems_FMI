package main

import (
	"bufio"
	"fmt"
	"os"
)

func startState(n int) []rune {
	s := make([]rune, 2*n+1)
	for i := 0; i < n; i++ {
		s[i] = '>'
	}
	s[n] = '_'
	for i := 0; i < n; i++ {
		s[n+1+i] = '<'
	}
	return s
}

func goalState(n int) string {
	s := make([]rune, 2*n+1)
	for i := 0; i < n; i++ {
		s[i] = '<'
	}
	s[n] = '_'
	for i := 0; i < n; i++ {
		s[n+1+i] = '>'
	}
	return string(s)
}

func cloneRunes(a []rune) []rune {
	b := make([]rune, len(a))
	copy(b, a)
	return b
}

func toStr(a []rune) string { return string(a) }

func inversions(state []rune) int {
	seenRight := 0
	inv := 0
	for _, ch := range state {
		if ch == '>' {
			seenRight++
		} else if ch == '<' {
			inv += seenRight
		}
	}
	return inv
}

func nextStates(state []rune) []string {
	n := len(state)
	var res []string

	// '>' to the right: jump first, then step
	for i, ch := range state {
		if ch != '>' {
			continue
		}
		// jump
		if i+2 < n && (state[i+1] == '>' || state[i+1] == '<') && state[i+2] == '_' {
			t := cloneRunes(state)
			t[i], t[i+2] = '_', '>'
			res = append(res, toStr(t))
		}
		// step
		if i+1 < n && state[i+1] == '_' {
			t := cloneRunes(state)
			t[i], t[i+1] = '_', '>'
			res = append(res, toStr(t))
		}
	}

	// '<' to the left: iterate right-to-left; jump first, then step
	for i := n - 1; i >= 0; i-- {
		if state[i] != '<' {
			continue
		}
		// jump
		if i-2 >= 0 && (state[i-1] == '>' || state[i-1] == '<') && state[i-2] == '_' {
			t := cloneRunes(state)
			t[i], t[i-2] = '_', '<'
			res = append(res, toStr(t))
		}
		// step
		if i-1 >= 0 && state[i-1] == '_' {
			t := cloneRunes(state)
			t[i], t[i-1] = '_', '<'
			res = append(res, toStr(t))
		}
	}
	return res
}

func solveDFS(n int) ([]string, bool) {
	start := startState(n)
	goal := goalState(n)
	depthLimit := n*n + 2*n

	path := []string{toStr(start)}
	visited := map[string]bool{toStr(start): true}

	var dfs func(state []rune, depth int) bool
	dfs = func(state []rune, depth int) bool {
		cur := toStr(state)
		if cur == goal {
			return true
		}
		if depth >= depthLimit {
			return false
		}

		inv := inversions(state)
		minMovesNeeded := (inv + 1) / 2
		if depth+minMovesNeeded > depthLimit {
			return false
		}

		for _, ns := range nextStates(state) {
			if visited[ns] {
				continue
			}
			visited[ns] = true
			path = append(path, ns)
			if dfs([]rune(ns), depth+1) {
				return true
			}
			path = path[:len(path)-1]
			// keep visited marked for stronger pruning
		}
		return false
	}

	ok := dfs(start, 0)
	return path, ok
}

func main() {
	in := bufio.NewReader(os.Stdin)
	var N int
	if _, err := fmt.Fscan(in, &N); err != nil {
		fmt.Fprintln(os.Stderr, "Please provide N (integer).")
		os.Exit(1)
	}

	var path []string

	if p, ok := solveDFS(N); ok {
		path = p
	}

	w := bufio.NewWriter(os.Stdout)
	for _, s := range path {
		fmt.Fprintln(w, s)
	}
	w.Flush()
}
