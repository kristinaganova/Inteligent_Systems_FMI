package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func createState(n int, a, b byte) []byte {
	s := make([]byte, 2*n+1)
	for i := 0; i < n; i++ {
		s[i] = a
	}
	s[n] = '_'
	for i := 0; i < n; i++ {
		s[n+1+i] = b
	}
	return s
}

func inversions(state []byte) int {
	seenR, inv := 0, 0
	for _, ch := range state {
		switch ch {
		case '>':
			seenR++
		case '<':
			inv += seenR
		}
	}
	return inv
}

func symIdx(b byte) int {
	if b == '>' {
		return 0
	}
	if b == '_' {
		return 1
	}
	return 2
}

func zobristTable(n int) [][3]uint64 {
	L := 2*n + 1
	tab := make([][3]uint64, L)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < L; i++ {
		tab[i][0] = r.Uint64() // '>'
		tab[i][1] = r.Uint64() // '_'
		tab[i][2] = r.Uint64() // '<'
	}
	return tab
}

func hashState(s []byte, z [][3]uint64) uint64 {
	var h uint64
	for i, b := range s {
		h ^= z[i][symIdx(b)]
	}
	return h
}

func search(n int, timeOnly bool) []string {
	state := createState(n, '>', '<')
	depthLimit := n*n + 2*n
	L := len(state)
	blank := n
	inv := inversions(state)

	z := zobristTable(n)
	hb := hashState(state, z)
	visited := map[uint64]struct{}{hb: {}}

	var path []string
	if !timeOnly {
		path = make([]string, 0, depthLimit+1)
		path = append(path, string(state))
	}

	var dfs func(depth, blank, inv int, hb uint64) bool
	dfs = func(depth, blank, inv int, hb uint64) bool {

		if inv == 0 && blank == n {
			return true
		}
		if depth >= depthLimit {
			return false
		}
		if depth+(inv+1)/2 > depthLimit {
			return false
		}

		// '>' jump
		if blank-2 >= 0 && state[blank-2] == '>' && state[blank-1] != '_' {
			delta := 0
			if state[blank-1] == '<' {
				delta = -1
			}
			i := blank - 2
			oldA, oldB := state[blank], state[i]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			if _, ok := visited[hb]; !ok {
				visited[hb] = struct{}{}
				if !timeOnly {
					path = append(path, string(state))
				}
				if dfs(depth+1, i, inv+delta, hb) {
					return true
				}
				if !timeOnly {
					path = path[:len(path)-1]
				}
			}
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
		}

		// '>' step
		if blank-1 >= 0 && state[blank-1] == '>' {
			i := blank - 1
			oldA, oldB := state[blank], state[i]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			if _, ok := visited[hb]; !ok {
				visited[hb] = struct{}{}
				if !timeOnly {
					path = append(path, string(state))
				}
				if dfs(depth+1, i, inv, hb) {
					return true
				}
				if !timeOnly {
					path = path[:len(path)-1]
				}
			}
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
		}

		// '<' jump
		if blank+2 < L && state[blank+2] == '<' && state[blank+1] != '_' {
			delta := 0
			if state[blank+1] == '>' {
				delta = -1
			}
			i := blank + 2
			oldA, oldB := state[blank], state[i]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			if _, ok := visited[hb]; !ok {
				visited[hb] = struct{}{}
				if !timeOnly {
					path = append(path, string(state))
				}
				if dfs(depth+1, i, inv+delta, hb) {
					return true
				}
				if !timeOnly {
					path = path[:len(path)-1]
				}
			}
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
		}

		// '<' step
		if blank+1 < L && state[blank+1] == '<' {
			i := blank + 1
			oldA, oldB := state[blank], state[i]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			if _, ok := visited[hb]; !ok {
				visited[hb] = struct{}{}
				if !timeOnly {
					path = append(path, string(state))
				}
				if dfs(depth+1, i, inv, hb) {
					return true
				}
				if !timeOnly {
					path = path[:len(path)-1]
				}
			}
			hb ^= z[blank][symIdx(state[blank])] ^ z[i][symIdx(state[i])]
			state[blank], state[i] = state[i], state[blank]
			hb ^= z[blank][symIdx(oldA)] ^ z[i][symIdx(oldB)]
		}
		return false
	}

	dfs(0, blank, inv, hb)
	return path
}

func main() {
	in := bufio.NewReader(os.Stdin)
	var N int
	fmt.Fscan(in, &N)

	timeOnly := os.Getenv("FMI_TIME_ONLY") == "1"
	start := time.Now()
	path := search(N, timeOnly)
	elapsed := time.Since(start)

	if timeOnly {
		fmt.Printf("# TIMES_MS: alg=%d\n", elapsed.Milliseconds())
		return
	}

	w := bufio.NewWriterSize(os.Stdout, 1<<20)
	for _, s := range path {
		w.WriteString(s)
		w.WriteByte('\n')
	}
	w.Flush()
}
