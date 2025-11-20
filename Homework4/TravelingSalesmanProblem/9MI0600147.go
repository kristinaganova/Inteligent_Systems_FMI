package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type City struct {
	Index int
	Name  string
	X     float64
	Y     float64
}

type Genome struct {
	Route    []int
	Distance float64
	Fitness  float64
}

type TravelingSalesmanProblem struct {
	Cities              []City
	PopulationSize      int
	Population          []Genome
	Generations         int
	MutationProbability float64
}

func createRandomGenome(cities []City) Genome {
	n := len(cities)
	route := rand.Perm(n)
	g := Genome{Route: route}
	g.Distance = calculateDistance(g.Route, cities)
	g.Fitness = 1.0 / (g.Distance + 1.0)
	return g
}

func createTravelingSalesmanProblem(cities []City, populationSize, generations int, mutationProbability float64) *TravelingSalesmanProblem {
	tsp := &TravelingSalesmanProblem{
		Cities:              cities,
		PopulationSize:      populationSize,
		Generations:         generations,
		MutationProbability: mutationProbability,
		Population:          make([]Genome, populationSize),
	}
	for i := 0; i < populationSize; i++ {
		tsp.Population[i] = createRandomGenome(tsp.Cities)
	}

	sort.Slice(tsp.Population, func(i, j int) bool {
		return tsp.Population[i].Fitness > tsp.Population[j].Fitness
	})
	return tsp
}

func calculateDistance(route []int, cities []City) float64 {
	totalDistance := 0.0
	if len(route) == 0 {
		return totalDistance
	}
	for i := 0; i < len(route)-1; i++ {
		fromCity := cities[route[i]]
		toCity := cities[route[i+1]]
		dx := toCity.X - fromCity.X
		dy := toCity.Y - fromCity.Y
		totalDistance += math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))
	}

	return totalDistance
}

func (tsp *TravelingSalesmanProblem) tournamentSelection(k int) Genome {
	if len(tsp.Population) == 0 {
		return Genome{}
	}
	best := tsp.Population[rand.Intn(len(tsp.Population))]
	for i := 1; i < k; i++ {
		candidate := tsp.Population[rand.Intn(len(tsp.Population))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	return best
}

func (tsp *TravelingSalesmanProblem) crossover(parent1, parent2 Genome) []Genome {
	n := len(parent1.Route)
	if n == 0 {
		return []Genome{{}, {}}
	}
	// Ordered crossover (OX)
	i := rand.Intn(n)
	j := rand.Intn(n)
	if i > j {
		i, j = j, i
	}
	child1Route := make([]int, n)
	child2Route := make([]int, n)
	for idx := 0; idx < n; idx++ {
		child1Route[idx] = -1
		child2Route[idx] = -1
	}
	// copy slice
	for idx := i; idx <= j; idx++ {
		child1Route[idx] = parent1.Route[idx]
		child2Route[idx] = parent2.Route[idx]
	}
	fillFrom := func(child []int, donor []int) {
		pos := (j + 1) % n
		for _, gene := range donor {
			already := false
			for _, v := range child {
				if v == gene {
					already = true
					break
				}
			}
			if already {
				continue
			}
			// find next empty slot
			for child[pos] != -1 {
				pos = (pos + 1) % n
			}
			child[pos] = gene
			pos = (pos + 1) % n
		}
	}
	fillFrom(child1Route, parent2.Route)
	fillFrom(child2Route, parent1.Route)

	child1 := Genome{Route: child1Route}
	child1.Distance = calculateDistance(child1.Route, tsp.Cities)
	child1.Fitness = 1.0 / (child1.Distance + 1.0)

	child2 := Genome{Route: child2Route}
	child2.Distance = calculateDistance(child2.Route, tsp.Cities)
	child2.Fitness = 1.0 / (child2.Distance + 1.0)

	return []Genome{child1, child2}
}

func (tsp *TravelingSalesmanProblem) mutate(g Genome) Genome {
	if len(g.Route) <= 1 {
		return g
	}
	if rand.Float64() < tsp.MutationProbability {
		i := rand.Intn(len(g.Route))
		j := rand.Intn(len(g.Route))
		g.Route[i], g.Route[j] = g.Route[j], g.Route[i]
		g.Distance = calculateDistance(g.Route, tsp.Cities)
		g.Fitness = 1.0 / (g.Distance + 1.0)
	}
	return g
}

func (tsp *TravelingSalesmanProblem) evolvePopulationWithElitism(eliteCount int) {
	if eliteCount < 0 {
		eliteCount = 0
	}
	if eliteCount > tsp.PopulationSize {
		eliteCount = tsp.PopulationSize
	}

	newPopulation := make([]Genome, 0, tsp.PopulationSize)

	for i := 0; i < eliteCount; i++ {
		newPopulation = append(newPopulation, tsp.Population[i])
	}

	for len(newPopulation) < tsp.PopulationSize {
		parent1 := tsp.tournamentSelection(3)
		parent2 := tsp.tournamentSelection(3)

		children := tsp.crossover(parent1, parent2)
		children[0] = tsp.mutate(children[0])
		children[1] = tsp.mutate(children[1])

		newPopulation = append(newPopulation, children...)
	}

	if len(newPopulation) > tsp.PopulationSize {
		newPopulation = newPopulation[:tsp.PopulationSize]
	}

	tsp.Population = newPopulation

	sort.Slice(tsp.Population, func(i, j int) bool {
		return tsp.Population[i].Fitness > tsp.Population[j].Fitness
	})
}

func (tsp *TravelingSalesmanProblem) runAndLogEvolution() Genome {
	for i := 0; i < tsp.Generations; i++ {
		tsp.evolvePopulationWithElitism(2)

		if i%100 == 0 {
			fmt.Printf("Generation %d: Best Distance = %f\n", i, tsp.Population[0].Distance)
		}
	}
	fmt.Println(tsp.Population[0].Distance)
	fmt.Println()

	return tsp.Population[0]
}

func loadFiles(cities *[]City, fileNamesPath, fileCoordinatesPath string) error {
	if fileNamesPath == "" || fileCoordinatesPath == "" {
		return fmt.Errorf("file paths cannot be empty")
	}

	namesFile, err := os.Open(fileNamesPath)
	if err != nil {
		return fmt.Errorf("error opening names file: %w", err)
	}
	defer namesFile.Close()

	coordsFile, err := os.Open(fileCoordinatesPath)
	if err != nil {
		return fmt.Errorf("error opening coordinates file: %w", err)
	}
	defer coordsFile.Close()

	namesScanner := bufio.NewScanner(namesFile)
	coordsScanner := bufio.NewScanner(coordsFile)

	index := 0
	for namesScanner.Scan() && coordsScanner.Scan() {
		name := strings.TrimSpace(namesScanner.Text())
		coordinates := strings.TrimSpace(coordsScanner.Text())
		if name == "" || coordinates == "" {
			continue
		}
		parts := strings.Split(coordinates, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid coordinates line: %s", coordinates)
		}
		x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return fmt.Errorf("invalid x in coordinates: %w", err)
		}
		y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return fmt.Errorf("invalid y in coordinates: %w", err)
		}
		*cities = append(*cities, City{
			Index: index,
			Name:  name,
			X:     x,
			Y:     y,
		})
		index++
	}

	if err := namesScanner.Err(); err != nil {
		return fmt.Errorf("error reading names file: %w", err)
	}
	if err := coordsScanner.Err(); err != nil {
		return fmt.Errorf("error reading coordinates file: %w", err)
	}

	return nil
}

func generateRandomCities(cities *[]City, count int) {
	for i := 0; i < count; i++ {
		*cities = append(*cities, City{
			Index: i,
			Name:  fmt.Sprintf("City%d", i+1),
			X:     rand.Float64() * 1000,
			Y:     rand.Float64() * 1000,
		})
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	reader := bufio.NewReader(os.Stdin)
	inputLine, _ := reader.ReadString('\n')
	input := strings.TrimSpace(inputLine)

	startTime := time.Now()

	var cities []City

	if input == "UK12" {
		namesPath := filepath.Join("resource", "uk12_name.csv")
		coordsPath := filepath.Join("resource", "uk12_xy.csv")
		if err := loadFiles(&cities, namesPath, coordsPath); err != nil {
			fmt.Println("Error loading cities:", err)
			return
		}
	} else {
		countOfCities, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input")
			return
		}
		if countOfCities > 100 {
			fmt.Println("Count of cities cannot be more than 100")
			return
		}
		generateRandomCities(&cities, countOfCities)
	}

	tsp := createTravelingSalesmanProblem(cities, 350, 2500, 0.5)

	bestGenome := tsp.runAndLogEvolution()

	path := make([]string, 0, len(cities))
	for i := 0; i < len(cities); i++ {
		currentCity := cities[bestGenome.Route[i]]
		if input == "UK12" {
			path = append(path, currentCity.Name)
		} else {
			path = append(path, fmt.Sprintf("(%f, %f)", currentCity.X, currentCity.Y))
		}
	}

	fmt.Println(strings.Join(path, " -> "))
	fmt.Println(bestGenome.Distance)

	elapsed := time.Since(startTime).Seconds()
	fmt.Printf("%.2f\n", elapsed)
}
