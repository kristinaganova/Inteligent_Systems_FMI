package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Dataset struct {
	X [][]string
	y []string
}

type NBModel struct {
	Classes       []string
	PriorLog      map[string]float64              // log P(c)
	CondLog       map[string][]map[string]float64 // class -> featureIndex -> value -> log P(x=v|c)
	DefaultLog    map[string][]float64            // class -> featureIndex -> log P(unseen|c)
	FeatureVocabs []map[string]struct{}           // featureIndex -> set(values)
	Lambda        float64
	NumFeatures   int
}

func fetchCongressionalVoting() (Dataset, error) {
	url := "https://archive.ics.uci.edu/ml/machine-learning-databases/voting-records/house-votes-84.data"

	resp, err := http.Get(url)
	if err != nil {
		return Dataset{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Dataset{}, fmt.Errorf("http status %d", resp.StatusCode)
	}

	r := csv.NewReader(resp.Body)
	r.Comma = ','
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	var X [][]string
	var y []string

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Dataset{}, err
		}
		if len(rec) != 17 {
			continue
		}
		label := strings.TrimSpace(rec[0])
		feat := make([]string, 16)
		for i := 0; i < 16; i++ {
			v := strings.TrimSpace(rec[i+1])
			if v == "" {
				v = "?"
			}
			feat[i] = v
		}
		X = append(X, feat)
		y = append(y, label)
	}

	return Dataset{X: X, y: y}, nil
}

func uniqueStrings(a []string) []string {
	m := map[string]struct{}{}
	for _, s := range a {
		m[s] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func stratifiedSplit(X [][]string, y []string, trainRatio float64, seed int64) (Xtr [][]string, ytr []string, Xte [][]string, yte []string) {
	rng := rand.New(rand.NewSource(seed))
	classes := uniqueStrings(y)

	idxByClass := map[string][]int{}
	for i, c := range y {
		idxByClass[c] = append(idxByClass[c], i)
	}

	for _, c := range classes {
		idxs := idxByClass[c]
		rng.Shuffle(len(idxs), func(i, j int) { idxs[i], idxs[j] = idxs[j], idxs[i] })

		split := int(math.Floor(float64(len(idxs)) * trainRatio))
		for _, id := range idxs[:split] {
			Xtr = append(Xtr, X[id])
			ytr = append(ytr, y[id])
		}
		for _, id := range idxs[split:] {
			Xte = append(Xte, X[id])
			yte = append(yte, y[id])
		}
	}

	rng.Shuffle(len(Xtr), func(i, j int) {
		Xtr[i], Xtr[j] = Xtr[j], Xtr[i]
		ytr[i], ytr[j] = ytr[j], ytr[i]
	})
	rng.Shuffle(len(Xte), func(i, j int) {
		Xte[i], Xte[j] = Xte[j], Xte[i]
		yte[i], yte[j] = yte[j], yte[i]
	})

	return
}

func computeColumnModesTrain(Xtr [][]string) []string {
	modes := make([]string, 16)
	for j := 0; j < 16; j++ {
		count := map[string]int{}
		for i := range Xtr {
			v := Xtr[i][j]
			if v == "?" {
				continue
			}
			count[v]++
		}
		bestV := "?"
		bestC := -1
		for v, c := range count {
			if c > bestC {
				bestC = c
				bestV = v
			}
		}
		modes[j] = bestV
	}
	return modes
}

func applyImputation(X [][]string, modes []string) [][]string {
	out := make([][]string, len(X))
	for i := range X {
		row := make([]string, 16)
		copy(row, X[i])
		for j := 0; j < 16; j++ {
			if row[j] == "?" {
				row[j] = modes[j]
			}
		}
		out[i] = row
	}
	return out
}

func buildFeatureVocabs(X [][]string) []map[string]struct{} {
	vocabs := make([]map[string]struct{}, 16)
	for j := 0; j < 16; j++ {
		vocabs[j] = map[string]struct{}{}
		for i := range X {
			vocabs[j][X[i][j]] = struct{}{}
		}
	}
	return vocabs
}

func (m *NBModel) Fit(X [][]string, y []string) {
	m.NumFeatures = 16
	m.Classes = uniqueStrings(y)
	m.PriorLog = map[string]float64{}
	m.CondLog = map[string][]map[string]float64{}
	m.DefaultLog = map[string][]float64{}
	m.FeatureVocabs = buildFeatureVocabs(X)

	n := float64(len(y))
	for _, c := range m.Classes {
		cnt := 0
		for _, yc := range y {
			if yc == c {
				cnt++
			}
		}
		m.PriorLog[c] = math.Log(float64(cnt) / n)
	}

	for _, c := range m.Classes {
		var Xc [][]string
		for i := range X {
			if y[i] == c {
				Xc = append(Xc, X[i])
			}
		}
		Nc := float64(len(Xc))

		m.CondLog[c] = make([]map[string]float64, 16)
		m.DefaultLog[c] = make([]float64, 16)

		for j := 0; j < 16; j++ {
			m.CondLog[c][j] = map[string]float64{}
			counts := map[string]int{}
			for i := range Xc {
				counts[Xc[i][j]]++
			}
			V := float64(len(m.FeatureVocabs[j]))
			denom := Nc + m.Lambda*V

			m.DefaultLog[c][j] = math.Log((m.Lambda) / denom)

			for v := range m.FeatureVocabs[j] {
				num := float64(counts[v]) + m.Lambda
				m.CondLog[c][j][v] = math.Log(num / denom)
			}
		}
	}
}

func (m *NBModel) PredictOne(x []string) string {
	bestC := ""
	bestScore := math.Inf(-1)

	for _, c := range m.Classes {
		score := m.PriorLog[c]
		for j := 0; j < 16; j++ {
			if lp, ok := m.CondLog[c][j][x[j]]; ok {
				score += lp
			} else {
				score += m.DefaultLog[c][j]
			}
		}
		if score > bestScore {
			bestScore = score
			bestC = c
		}
	}
	return bestC
}

func accuracy(model *NBModel, X [][]string, y []string) float64 {
	correct := 0
	for i := range X {
		if model.PredictOne(X[i]) == y[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(y))
}

func stratifiedKFolds(X [][]string, y []string, k int, seed int64) []([]int) {
	rng := rand.New(rand.NewSource(seed))
	classes := uniqueStrings(y)
	idxByClass := map[string][]int{}
	for i, c := range y {
		idxByClass[c] = append(idxByClass[c], i)
	}

	folds := make([][]int, k)
	for _, c := range classes {
		idxs := idxByClass[c]
		rng.Shuffle(len(idxs), func(i, j int) { idxs[i], idxs[j] = idxs[j], idxs[i] })

		for i, id := range idxs {
			folds[i%k] = append(folds[i%k], id)
		}
	}
	return folds
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	m := 0.0
	for _, v := range xs {
		m += v
	}
	m /= float64(len(xs))
	if len(xs) == 1 {
		return m, 0
	}
	var s float64
	for _, v := range xs {
		d := v - m
		s += d * d
	}
	s /= float64(len(xs) - 1)
	return m, math.Sqrt(s)
}

type Result struct {
	Lambda   float64
	TrainAcc float64
	CVFolds  []float64
	CVMean   float64
	CVStd    float64
	TestAcc  float64
}

func main() {
	in := bufio.NewReader(os.Stdin)
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		fmt.Println("Expected input 0 or 1.")
		return
	}
	mode := 0
	if line == "1" {
		mode = 1
	} else if line != "0" {
		fmt.Println("Expected input 0 or 1.")
		return
	}

	ds, err := fetchCongressionalVoting()
	if err != nil {
		fmt.Println("Error fetching dataset:", err)
		return
	}

	seed := time.Now().UnixNano()
	Xtr, ytr, Xte, yte := stratifiedSplit(ds.X, ds.y, 0.8, seed)

	if mode == 1 {
		modes := computeColumnModesTrain(Xtr)
		Xtr = applyImputation(Xtr, modes)
		Xte = applyImputation(Xte, modes)
	}

	fmt.Printf("Chosen input mode: %d\n\n", mode)

	lambdas := []float64{0.0, 0.1, 0.5, 1.0, 2.0}
	var results []Result

	for _, lam := range lambdas {
		model := &NBModel{Lambda: lam}
		model.Fit(Xtr, ytr)

		trainAcc := accuracy(model, Xtr, ytr)

		folds := stratifiedKFolds(Xtr, ytr, 10, seed+12345)
		var foldAccs []float64

		for i := 0; i < 10; i++ {
			testIdx := map[int]struct{}{}
			for _, id := range folds[i] {
				testIdx[id] = struct{}{}
			}

			var X_train [][]string
			var y_train []string
			var X_valid [][]string
			var y_valid []string

			for idx := range Xtr {
				if _, ok := testIdx[idx]; ok {
					X_valid = append(X_valid, Xtr[idx])
					y_valid = append(y_valid, ytr[idx])
				} else {
					X_train = append(X_train, Xtr[idx])
					y_train = append(y_train, ytr[idx])
				}
			}

			m := &NBModel{Lambda: lam}
			m.Fit(X_train, y_train)
			foldAccs = append(foldAccs, accuracy(m, X_valid, y_valid))
		}

		cvMean, cvStd := meanStd(foldAccs)
		testAcc := accuracy(model, Xte, yte)

		results = append(results, Result{
			Lambda: lam, TrainAcc: trainAcc, CVFolds: foldAccs, CVMean: cvMean, CVStd: cvStd, TestAcc: testAcc,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].CVMean == results[j].CVMean {
			return results[i].TestAcc > results[j].TestAcc
		}
		return results[i].CVMean > results[j].CVMean
	})
	best := results[0]

	fmt.Printf("===== BEST λ selected by CV mean: λ = %.2f =====\n\n", best.Lambda)

	fmt.Println("1. Train Set Accuracy:")
	fmt.Printf("    Accuracy: %.2f%%\n\n", best.TrainAcc*100)

	fmt.Println("10-Fold Cross-Validation Results:\n")
	for i, acc := range best.CVFolds {
		fmt.Printf("    Accuracy Fold %d: %.2f%%\n", i+1, acc*100)
	}
	fmt.Printf("\n    Average Accuracy: %.2f%%\n", best.CVMean*100)
	fmt.Printf("    Standard Deviation: %.2f%%\n\n", best.CVStd*100)

	fmt.Println("2. Test Set Accuracy:")
	fmt.Printf("    Accuracy: %.2f%%\n\n", best.TestAcc*100)

	fmt.Println("----- Summary over λ (Train / CV mean±std / Test) -----")
	for _, r := range results {
		fmt.Printf("λ=%-4.1f | Train=%6.2f%% | CV=%6.2f%% ± %5.2f%% | Test=%6.2f%%\n",
			r.Lambda, r.TrainAcc*100, r.CVMean*100, r.CVStd*100, r.TestAcc*100)
	}
}
