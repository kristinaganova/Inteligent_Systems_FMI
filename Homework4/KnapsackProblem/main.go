package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"
)

type Item struct {
	Index  int
	Weight int
	Value  int
}

type Candidate struct {
	Genes []int
	Value float64
}

type KnapsackSolver struct {
	Items          []Item
	Capacity       int
	PopSize        int
	MaxGenerations int
	Population     []Candidate
	MutationRate   float64
	BestValues     []float64
}

func (c *Candidate) Evaluate(items []Item, capacity int) {
	totalWeight := 0
	totalValue := 0.0
	for i, gene := range c.Genes {
		if gene == 1 {
			totalWeight += items[i].Weight
			totalValue += float64(items[i].Value)
		}
	}
	if totalWeight <= capacity {
		c.Value = totalValue
	} else {
		c.Value = 0
	}
}

func (ks *KnapsackSolver) randomFeasibleGenes() []int {
	genes := make([]int, len(ks.Items))
	order := rand.Perm(len(ks.Items))
	totalWeight := 0
	for _, idx := range order {
		if totalWeight+ks.Items[idx].Weight <= ks.Capacity {
			genes[idx] = 1
			totalWeight += ks.Items[idx].Weight
		}
	}
	return genes
}

func (ks *KnapsackSolver) InitPopulation() {
	ks.Population = make([]Candidate, ks.PopSize)
	for i := 0; i < ks.PopSize; i++ {
		genes := ks.randomFeasibleGenes()
		c := Candidate{Genes: genes}
		c.Evaluate(ks.Items, ks.Capacity)
		ks.Population[i] = c
	}
	sort.Slice(ks.Population, func(i, j int) bool {
		return ks.Population[i].Value > ks.Population[j].Value
	})
	ks.BestValues = append(ks.BestValues, ks.Population[0].Value)
}

func (ks *KnapsackSolver) tournamentSelect(size int) Candidate {
	best := ks.Population[rand.Intn(ks.PopSize)]
	for i := 1; i < size; i++ {
		other := ks.Population[rand.Intn(ks.PopSize)]
		if other.Value > best.Value {
			best = other
		}
	}
	return best
}

func crossoverTwoPoint(p1, p2 Candidate) (Candidate, Candidate) {
	n := len(p1.Genes)
	if n <= 1 {
		return p1, p2
	}
	p1i := rand.Intn(n)
	p2i := rand.Intn(n)
	if p1i > p2i {
		p1i, p2i = p2i, p1i
	}
	if p1i == p2i {
		return p1, p2
	}
	c1 := Candidate{Genes: make([]int, n)}
	c2 := Candidate{Genes: make([]int, n)}

	copy(c1.Genes[p1i:p2i], p1.Genes[p1i:p2i])
	copy(c2.Genes[p1i:p2i], p2.Genes[p1i:p2i])

	for i := 0; i < n; i++ {
		if i < p1i || i >= p2i {
			c1.Genes[i] = p2.Genes[i]
			c2.Genes[i] = p1.Genes[i]
		}
	}
	return c1, c2
}

func (ks *KnapsackSolver) totalWeight(genes []int) int {
	sum := 0
	for i, g := range genes {
		if g == 1 {
			sum += ks.Items[i].Weight
		}
	}
	return sum
}

func (ks *KnapsackSolver) mutateCandidate(c Candidate) Candidate {
	currWeight := ks.totalWeight(c.Genes)
	for i := range c.Genes {
		if rand.Float64() < ks.MutationRate {
			if c.Genes[i] == 0 {
				if currWeight+ks.Items[i].Weight <= ks.Capacity {
					c.Genes[i] = 1
					currWeight += ks.Items[i].Weight
				}
			} else {
				c.Genes[i] = 0
				currWeight -= ks.Items[i].Weight
			}
		}
	}
	c.Evaluate(ks.Items, ks.Capacity)
	return c
}

func (ks *KnapsackSolver) evolveStep(eliteSize, tournamentSize int) {
	sort.Slice(ks.Population, func(i, j int) bool {
		return ks.Population[i].Value > ks.Population[j].Value
	})

	next := make([]Candidate, 0, ks.PopSize)
	for i := 0; i < eliteSize && i < len(ks.Population); i++ {
		next = append(next, ks.Population[i])
	}

	for len(next) < ks.PopSize {
		p1 := ks.tournamentSelect(tournamentSize)
		p2 := ks.tournamentSelect(tournamentSize)
		c1, c2 := crossoverTwoPoint(p1, p2)
		c1 = ks.mutateCandidate(c1)
		if len(next) < ks.PopSize {
			next = append(next, c1)
		}
		c2 = ks.mutateCandidate(c2)
		if len(next) < ks.PopSize {
			next = append(next, c2)
		}
	}

	sort.Slice(next, func(i, j int) bool {
		return next[i].Value > next[j].Value
	})
	ks.Population = next
	ks.BestValues = append(ks.BestValues, ks.Population[0].Value)
}

func (ks *KnapsackSolver) Run(eliteSize, tournamentSize int) Candidate {
	for gen := 0; gen < ks.MaxGenerations; gen++ {
		ks.evolveStep(eliteSize, tournamentSize)
	}
	return ks.Population[0]
}

func main() {
	rand.NewSource(time.Now().UnixNano())

	measureTime := flag.Bool("time", false, "print elapsed time to stderr")
	flag.Parse()

	start := time.Now()

	in := bufio.NewReader(os.Stdin)
	var capacity, n int
	if _, err := fmt.Fscan(in, &capacity, &n); err != nil {
		return
	}

	items := make([]Item, n)
	for i := 0; i < n; i++ {
		var w, v int
		fmt.Fscan(in, &w, &v)
		items[i] = Item{Index: i, Weight: w, Value: v}
	}

	popSize := 300
	generations := 1500
	if n > 300 {
		popSize = 200
		generations = 1200
	}
	if n > 1000 {
		popSize = 150
		generations = 800
	}
	mutationRate := 0.03

	solver := KnapsackSolver{
		Items:          items,
		Capacity:       capacity,
		PopSize:        popSize,
		MaxGenerations: generations,
		MutationRate:   mutationRate,
	}

	solver.InitPopulation()
	best := solver.Run(5, 3)

	if *measureTime {
		fmt.Fprintf(os.Stderr, "Elapsed: %.6f seconds\n", time.Since(start).Seconds())
	}

	progress := solver.BestValues
	if len(progress) == 0 {
		fmt.Println("0")
		fmt.Println()
		fmt.Println("0")
		return
	}

	lastIndex := len(progress) - 1
	for i := 0; i < 10; i++ {
		idx := i * lastIndex / 9
		fmt.Printf("%.0f\n", progress[idx])
	}

	fmt.Println()
	fmt.Printf("%.0f\n", best.Value)
}
