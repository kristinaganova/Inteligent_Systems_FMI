package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Point struct {
	Features []float64
	Class    string
}

type Dataset struct {
	Points []Point
}

type KDTree struct {
	Point     *Point
	Left      *KDTree
	Right     *KDTree
	Axis      int
	Dimension int
}

type Neighbor struct {
	Point    *Point
	Distance float64
}

func LoadData(filename string) (*Dataset, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	dataset := &Dataset{Points: make([]Point, 0)}

	for _, record := range records {
		if len(record) < 5 {
			continue
		}

		features := make([]float64, 4)
		for i := 0; i < 4; i++ {
			val, err := strconv.ParseFloat(strings.TrimSpace(record[i]), 64)
			if err != nil {
				continue
			}
			features[i] = val
		}

		class := strings.TrimSpace(record[4])
		dataset.Points = append(dataset.Points, Point{
			Features: features,
			Class:    class,
		})
	}

	return dataset, nil
}

func MinMaxNormalize(dataset *Dataset) {
	if len(dataset.Points) == 0 {
		return
	}

	dim := len(dataset.Points[0].Features)
	mins := make([]float64, dim)
	maxs := make([]float64, dim)

	for i := 0; i < dim; i++ {
		mins[i] = dataset.Points[0].Features[i]
		maxs[i] = dataset.Points[0].Features[i]
	}

	for _, point := range dataset.Points {
		for i := 0; i < dim; i++ {
			if point.Features[i] < mins[i] {
				mins[i] = point.Features[i]
			}
			if point.Features[i] > maxs[i] {
				maxs[i] = point.Features[i]
			}
		}
	}

	for i := range dataset.Points {
		for j := 0; j < dim; j++ {
			diff := maxs[j] - mins[j]
			if diff != 0 {
				dataset.Points[i].Features[j] = (dataset.Points[i].Features[j] - mins[j]) / diff
			} else {
				dataset.Points[i].Features[j] = 0
			}
		}
	}
}

func StratifiedSplit(dataset *Dataset, testSize float64, seed int64) (*Dataset, *Dataset) {
	r := rand.New(rand.NewSource(seed))

	classGroups := make(map[string][]Point)
	for _, p := range dataset.Points {
		classGroups[p.Class] = append(classGroups[p.Class], p)
	}

	trainPoints := make([]Point, 0, len(dataset.Points))
	testPoints := make([]Point, 0)

	classNames := make([]string, 0, len(classGroups))
	for c := range classGroups {
		classNames = append(classNames, c)
	}
	sort.Strings(classNames)

	for _, c := range classNames {
		points := classGroups[c]

		shuffled := make([]Point, len(points))
		copy(shuffled, points)
		r.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		n := len(shuffled)

		// If a class has only 1 sample, keep it in train to avoid empty train for that class.
		if n == 1 {
			trainPoints = append(trainPoints, shuffled[0])
			continue
		}

		nTest := int(math.Round(float64(n) * testSize))

		if nTest < 0 {
			nTest = 0
		}
		if nTest > n {
			nTest = n
		}

		if testSize > 0 && nTest == 0 {
			nTest = 1
		}

		// Ensure at least 1 train sample
		if nTest >= n {
			nTest = n - 1
		}

		testPoints = append(testPoints, shuffled[:nTest]...)
		trainPoints = append(trainPoints, shuffled[nTest:]...)
	}

	return &Dataset{Points: trainPoints}, &Dataset{Points: testPoints}
}

func EuclideanDistance(p1, p2 []float64) float64 {
	sum := 0.0
	for i := 0; i < len(p1) && i < len(p2); i++ {
		diff := p1[i] - p2[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func BuildKDTree(points []Point, depth int) *KDTree {
	if len(points) == 0 {
		return nil
	}

	if len(points) == 1 {
		return &KDTree{
			Point:     &points[0],
			Left:      nil,
			Right:     nil,
			Axis:      depth % len(points[0].Features),
			Dimension: len(points[0].Features),
		}
	}

	dim := len(points[0].Features)
	axis := depth % dim

	sortedPoints := make([]Point, len(points))
	copy(sortedPoints, points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].Features[axis] < sortedPoints[j].Features[axis]
	})

	median := len(sortedPoints) / 2

	node := &KDTree{
		Point:     &sortedPoints[median],
		Axis:      axis,
		Dimension: dim,
	}

	node.Left = BuildKDTree(sortedPoints[:median], depth+1)
	node.Right = BuildKDTree(sortedPoints[median+1:], depth+1)

	return node
}

func KNNQuery(tree *KDTree, query []float64, k int) []Neighbor {
	neighbors := make([]Neighbor, 0)
	knnQueryRecursive(tree, query, k, &neighbors)
	return neighbors
}

func knnQueryRecursive(node *KDTree, query []float64, k int, neighbors *[]Neighbor) {
	if node == nil {
		return
	}

	dist := EuclideanDistance(query, node.Point.Features)

	*neighbors = append(*neighbors, Neighbor{Point: node.Point, Distance: dist})

	if len(*neighbors) > k {
		sort.Slice(*neighbors, func(i, j int) bool {
			return (*neighbors)[i].Distance < (*neighbors)[j].Distance
		})
		*neighbors = (*neighbors)[:k]
	} else {
		sort.Slice(*neighbors, func(i, j int) bool {
			return (*neighbors)[i].Distance < (*neighbors)[j].Distance
		})
	}

	axis := node.Axis
	axisDist := query[axis] - node.Point.Features[axis]

	var near, far *KDTree
	if axisDist < 0 {
		near = node.Left
		far = node.Right
	} else {
		near = node.Right
		far = node.Left
	}

	knnQueryRecursive(near, query, k, neighbors)

	maxDist := math.MaxFloat64
	if len(*neighbors) > 0 {
		maxDist = (*neighbors)[len(*neighbors)-1].Distance
	}

	if len(*neighbors) < k || math.Abs(axisDist) < maxDist {
		knnQueryRecursive(far, query, k, neighbors)
	}
}

func PredictClass(tree *KDTree, query []float64, k int) string {
	neighbors := KNNQuery(tree, query, k)

	votes := make(map[string]int)
	for _, neighbor := range neighbors {
		votes[neighbor.Point.Class]++
	}

	maxVotes := 0
	predictedClass := ""
	for class, count := range votes {
		if count > maxVotes {
			maxVotes = count
			predictedClass = class
		}
	}

	return predictedClass
}

func CalculateAccuracy(dataset *Dataset, tree *KDTree, k int) float64 {
	correct := 0
	for _, point := range dataset.Points {
		predicted := PredictClass(tree, point.Features, k)
		if predicted == point.Class {
			correct++
		}
	}
	return float64(correct) / float64(len(dataset.Points)) * 100.0
}

func CrossValidation(trainData *Dataset, k int, seed int64) (float64, float64) {
	r := rand.New(rand.NewSource(seed))

	points := make([]Point, len(trainData.Points))
	copy(points, trainData.Points)
	r.Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})

	foldSize := len(points) / 10
	accuracies := make([]float64, 10)

	for fold := 0; fold < 10; fold++ {
		start := fold * foldSize
		end := start + foldSize

		validationPoints := points[start:end]
		trainingPoints := make([]Point, 0)
		trainingPoints = append(trainingPoints, points[:start]...)
		trainingPoints = append(trainingPoints, points[end:]...)

		tree := BuildKDTree(trainingPoints, 0)

		correct := 0
		for _, point := range validationPoints {

			predicted := PredictClass(tree, point.Features, k)

			if predicted == point.Class {
				correct++
			}

		}

		accuracies[fold] = float64(correct) / float64(len(validationPoints)) * 100.0
		fmt.Printf("    Accuracy Fold %d: %.2f%%\n", fold+1, accuracies[fold])
	}

	avg := 0.0
	for _, acc := range accuracies {
		avg += acc
	}
	avg /= float64(len(accuracies))

	variance := 0.0
	for _, acc := range accuracies {
		variance += (acc - avg) * (acc - avg)
	}
	variance /= float64(len(accuracies))
	stdDev := math.Sqrt(variance)

	fmt.Printf("\n    Average Accuracy: %.2f%%\n", avg)
	fmt.Printf("    Standard Deviation: %.2f%%\n", stdDev)

	return avg, stdDev
}

func PlotAccuracyVsK(trainData, testData *Dataset, maxK int) {
	fmt.Println("\nAccuracy vs k (for plotting):")
	fmt.Println("k\tTrain Accuracy\tTest Accuracy")

	trainTree := BuildKDTree(trainData.Points, 0)

	for k := 1; k <= maxK; k += 2 {
		trainAcc := CalculateAccuracy(trainData, trainTree, k)
		testAcc := CalculateAccuracy(testData, trainTree, k)
		fmt.Printf("%d\t%.2f\t\t%.2f\n", k, trainAcc, testAcc)
	}
}

func main() {

	dataset, err := LoadData("iris.data")
	if err != nil {
		fmt.Printf("Error loading data: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d examples with %d features\n", len(dataset.Points), len(dataset.Points[0].Features))

	classCount := make(map[string]int)
	for _, point := range dataset.Points {
		classCount[point.Class]++
	}
	fmt.Printf("Class distribution: %v\n", classCount)

	MinMaxNormalize(dataset)
	fmt.Println("Data normalized using Min-Max normalization")

	trainData, testData := StratifiedSplit(dataset, 0.2, 56)
	fmt.Printf("\nTraining set: %d examples\n", len(trainData.Points))
	fmt.Printf("Test set: %d examples\n", len(testData.Points))

	fmt.Print("\nEnter value for k: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	var k int
	fmt.Sscanf(strings.TrimSpace(input), "%d", &k)

	if k <= 0 {
		fmt.Println("Error: k must be positive. Using k=1")
		k = 1
	}
	if k > len(trainData.Points) {
		fmt.Printf("Warning: k (%d) is larger than training set size (%d). Using k=%d\n", k, len(trainData.Points), len(trainData.Points))
		k = len(trainData.Points)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Results for k = %d\n", k)
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	trainTree := BuildKDTree(trainData.Points, 0)

	trainAccuracy := CalculateAccuracy(trainData, trainTree, k)
	fmt.Println("1. Train Set Accuracy:")
	fmt.Printf("   Accuracy: %.2f%%\n", trainAccuracy)

	fmt.Println("\n2. 10-Fold Cross-Validation Results:")
	CrossValidation(trainData, k, 42)

	testAccuracy := CalculateAccuracy(testData, trainTree, k)
	fmt.Println("\n3. Test Set Accuracy:")
	fmt.Printf("   Accuracy: %.2f%%\n", testAccuracy)
}
