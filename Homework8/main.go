package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Example map[string]string

type Dataset struct {
	Attrs     []string // feature attrs (без Class)
	ClassAttr string   // "Class"
	Data      []Example
	AttrVals  map[string][]string
}

type Node struct {
	IsLeaf        bool
	ClassLabel    string           // for leaf
	Attr          string           // for decision
	Children      map[string]*Node // value -> subtree
	MajorityClass string           // fallback for unseen values
	Depth         int
}

type PrePrune struct {
	UseN bool
	N    int
	UseK bool
	K    int
	UseG bool
	G    float64
}

type Config struct {
	Mode            string // "0","1","2"
	Pre             PrePrune
	UsePostE        bool
	Seed            int64
	TrainRatio      float64
	ValRatioInTrain float64 // used for post-pruning (split train into subtrain/val)
	Folds           int
	MissingPolicy   string // "mode_by_class"
}

func main() {
	arffPath := flag.String("data", "breast-cancer.arff", "path to breast-cancer.arff")
	input := flag.String("input", "2", `pruning input like: "0", "0 K", "1 E", "2 NKG E" (quotes recommended)`)
	seed := flag.Int64("seed", 42, "random seed")
	flag.Parse()

	cfg := defaultConfig()
	cfg.Seed = *seed
	parseInputPruning(*input, &cfg)

	ds, err := LoadARFF(*arffPath)
	if err != nil {
		fmt.Println("Error loading ARFF:", err)
		os.Exit(1)
	}

	imputeMissingModeByClass(&ds)

	rng := rand.New(rand.NewSource(cfg.Seed))
	train, test := stratifiedSplit(ds, cfg.TrainRatio, rng)

	model := trainWithOptionalPostPruning(train, cfg, rng)
	trainAcc := accuracy(model, train)

	foldAccs := stratifiedKFoldCV(train, cfg.Folds, cfg, rng)

	avg, std := meanStd(foldAccs)

	model2 := trainWithOptionalPostPruning(train, cfg, rng)
	testAcc := accuracy(model2, test)

	fmt.Printf("1. Train Set Accuracy:\n")
	fmt.Printf("    Accuracy: %.2f%%\n\n", trainAcc*100)

	fmt.Printf("10-Fold Cross-Validation Results:\n\n")
	for i, a := range foldAccs {
		fmt.Printf("    Accuracy Fold %d: %.2f%%\n", i+1, a*100)
	}
	fmt.Printf("\n    Average Accuracy: %.2f%%\n", avg*100)
	fmt.Printf("    Standard Deviation: %.2f%%\n\n", std*100)

	fmt.Printf("2. Test Set Accuracy:\n")
	fmt.Printf("    Accuracy: %.2f%%\n", testAcc*100)
}

func defaultConfig() Config {
	return Config{
		Mode:            "2",
		Seed:            42,
		TrainRatio:      0.8,
		ValRatioInTrain: 0.2, // used only if post-pruning enabled
		Folds:           10,
		MissingPolicy:   "mode_by_class",
		Pre: PrePrune{
			UseN: true, N: 10,
			UseK: true, K: 5,
			UseG: true, G: 0.1,
		},
		UsePostE: true,
	}
}

// input examples:
// "0" => all implemented pre-pruning (N,K,G), no post
// "0 K" => only K
// "1" => all implemented post-pruning (E), no pre
// "1 E" => only E
// "2" => both all
// "2 NKG E" => both, only specified subsets
func parseInputPruning(s string, cfg *Config) {
	parts := strings.Fields(strings.TrimSpace(s))
	if len(parts) == 0 {
		return
	}
	mode := parts[0]
	cfg.Mode = mode

	switch mode {
	case "0":
		cfg.Pre.UseN, cfg.Pre.UseK, cfg.Pre.UseG = true, true, true
		cfg.UsePostE = false
	case "1":
		cfg.Pre.UseN, cfg.Pre.UseK, cfg.Pre.UseG = false, false, false
		cfg.UsePostE = true
	case "2":
		cfg.Pre.UseN, cfg.Pre.UseK, cfg.Pre.UseG = true, true, true
		cfg.UsePostE = true
	default:
	}

	if len(parts) > 1 {
		if mode == "0" || mode == "2" {
			cfg.Pre.UseN, cfg.Pre.UseK, cfg.Pre.UseG = false, false, false
		}
		if mode == "1" || mode == "2" {
			cfg.UsePostE = false
		}

		for _, t := range parts[1:] {
			up := strings.ToUpper(t)
			if strings.ContainsAny(up, "NKG") && (mode == "0" || mode == "2") {
				if strings.Contains(up, "N") {
					cfg.Pre.UseN = true
				}
				if strings.Contains(up, "K") {
					cfg.Pre.UseK = true
				}
				if strings.Contains(up, "G") {
					cfg.Pre.UseG = true
				}
			}
			if strings.Contains(up, "E") && (mode == "1" || mode == "2") {
				cfg.UsePostE = true
			}
		}

		if (mode == "0" || mode == "2") && !(cfg.Pre.UseN || cfg.Pre.UseK || cfg.Pre.UseG) {
			cfg.Pre.UseN, cfg.Pre.UseK, cfg.Pre.UseG = true, true, true
		}
		if (mode == "1" || mode == "2") && !cfg.UsePostE {
			cfg.UsePostE = true
		}
	}
}

func LoadARFF(path string) (Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return Dataset{}, err
	}
	defer f.Close()

	ds := Dataset{
		AttrVals: make(map[string][]string),
	}

	sc := bufio.NewScanner(f)
	inData := false
	var allAttrs []string

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}
		low := strings.ToLower(line)
		if strings.HasPrefix(low, "@relation") {
			continue
		}
		if strings.HasPrefix(low, "@attribute") {
			toks := splitARFFAttribute(line)
			if len(toks) < 3 {
				return Dataset{}, fmt.Errorf("bad @attribute line: %s", line)
			}
			name := toks[1]
			spec := strings.Join(toks[2:], " ")
			allAttrs = append(allAttrs, name)
			if strings.Contains(spec, "{") && strings.Contains(spec, "}") {
				vals := parseBraceList(spec)
				ds.AttrVals[name] = vals
			}
			continue
		}
		if strings.HasPrefix(low, "@data") {
			inData = true
			continue
		}
		if inData {
			parts := splitCSVLine(line)
			if len(parts) != len(allAttrs) {
				return Dataset{}, fmt.Errorf("data row has %d cols, expected %d: %s", len(parts), len(allAttrs), line)
			}
			ex := make(Example)
			for i, a := range allAttrs {
				v := strings.TrimSpace(parts[i])
				if v == "?" {
					v = ""
				}
				ex[a] = v
			}
			ds.Data = append(ds.Data, ex)
		}
	}
	if err := sc.Err(); err != nil {
		return Dataset{}, err
	}
	if len(allAttrs) == 0 || len(ds.Data) == 0 {
		return Dataset{}, errors.New("no attributes or no data found")
	}

	ds.ClassAttr = allAttrs[len(allAttrs)-1]
	ds.Attrs = append([]string{}, allAttrs[:len(allAttrs)-1]...)
	return ds, nil
}

func splitARFFAttribute(line string) []string {
	return strings.Fields(line)
}

func parseBraceList(spec string) []string {
	l := strings.Index(spec, "{")
	r := strings.LastIndex(spec, "}")
	if l < 0 || r < 0 || r <= l {
		return nil
	}
	inside := spec[l+1 : r]
	raw := strings.Split(inside, ",")
	out := make([]string, 0, len(raw))
	for _, x := range raw {
		v := strings.TrimSpace(x)
		out = append(out, v)
	}
	return out
}

func splitCSVLine(line string) []string {
	return strings.Split(line, ",")
}

func imputeMissingModeByClass(ds *Dataset) {
	classCounts := make(map[string]int)
	counts := make(map[string]map[string]map[string]int)
	globalCounts := make(map[string]map[string]int)

	for _, ex := range ds.Data {
		c := ex[ds.ClassAttr]
		classCounts[c]++
		if _, ok := counts[c]; !ok {
			counts[c] = make(map[string]map[string]int)
		}
		for _, a := range ds.Attrs {
			v := ex[a]
			if v == "" {
				continue
			}
			if _, ok := counts[c][a]; !ok {
				counts[c][a] = make(map[string]int)
			}
			counts[c][a][v]++

			if _, ok := globalCounts[a]; !ok {
				globalCounts[a] = make(map[string]int)
			}
			globalCounts[a][v]++
		}
	}

	modeByClass := make(map[string]map[string]string)
	globalMode := make(map[string]string)

	for _, a := range ds.Attrs {
		globalMode[a] = argmax(globalCounts[a])
	}

	for c := range classCounts {
		modeByClass[c] = make(map[string]string)
		for _, a := range ds.Attrs {
			mode := argmax(counts[c][a])
			if mode == "" {
				mode = globalMode[a]
			}
			modeByClass[c][a] = mode
		}
	}

	for i := range ds.Data {
		c := ds.Data[i][ds.ClassAttr]
		for _, a := range ds.Attrs {
			if ds.Data[i][a] == "" {
				ds.Data[i][a] = modeByClass[c][a]
			}
		}
	}
}

func argmax(m map[string]int) string {
	best := ""
	bestN := -1
	for k, v := range m {
		if v > bestN {
			bestN = v
			best = k
		}
	}
	return best
}

func stratifiedSplit(ds Dataset, trainRatio float64, rng *rand.Rand) (Dataset, Dataset) {
	byClass := make(map[string][]Example)
	for _, ex := range ds.Data {
		c := ex[ds.ClassAttr]
		byClass[c] = append(byClass[c], ex)
	}

	var trainData, testData []Example
	for c, arr := range byClass {
		_ = c
		shuffled := append([]Example{}, arr...)
		rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		split := int(math.Round(float64(len(shuffled)) * trainRatio))
		if split < 1 {
			split = 1
		}
		if split > len(shuffled)-1 {
			split = len(shuffled) - 1
		}
		trainData = append(trainData, shuffled[:split]...)
		testData = append(testData, shuffled[split:]...)
	}

	train := ds
	test := ds
	train.Data = trainData
	test.Data = testData
	return train, test
}

func stratifiedKFoldCV(train Dataset, k int, cfg Config, rng *rand.Rand) []float64 {
	folds := make([][]Example, k)
	byClass := make(map[string][]Example)
	for _, ex := range train.Data {
		byClass[ex[train.ClassAttr]] = append(byClass[ex[train.ClassAttr]], ex)
	}
	for _, arr := range byClass {
		shuffled := append([]Example{}, arr...)
		rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		for i, ex := range shuffled {
			folds[i%k] = append(folds[i%k], ex)
		}
	}

	accs := make([]float64, 0, k)
	for i := 0; i < k; i++ {
		var tr, te []Example
		for j := 0; j < k; j++ {
			if j == i {
				te = append(te, folds[j]...)
			} else {
				tr = append(tr, folds[j]...)
			}
		}
		dsTr := train
		dsTe := train
		dsTr.Data = tr
		dsTe.Data = te

		model := trainWithOptionalPostPruning(dsTr, cfg, rng)
		accs = append(accs, accuracy(model, dsTe))
	}
	return accs
}

func trainWithOptionalPostPruning(train Dataset, cfg Config, rng *rand.Rand) *Node {
	if cfg.UsePostE {
		// split train -> subtrain/val (stratified)
		subtrain, val := stratifiedSplit(train, 1.0-cfg.ValRatioInTrain, rng)
		tree := buildID3(subtrain, cfg, 0)
		tree = reducedErrorPrune(tree, val, train.ClassAttr)
		return tree
	}
	return buildID3(train, cfg, 0)
}

func buildID3(ds Dataset, cfg Config, depth int) *Node {
	maj := majorityClass(ds.Data, ds.ClassAttr)

	if isPure(ds.Data, ds.ClassAttr) {
		return &Node{IsLeaf: true, ClassLabel: ds.Data[0][ds.ClassAttr], MajorityClass: maj, Depth: depth}
	}

	if (cfg.Mode == "0" || cfg.Mode == "2") && (cfg.Pre.UseN || cfg.Pre.UseK || cfg.Pre.UseG) {
		if cfg.Pre.UseN && depth >= cfg.Pre.N {
			return &Node{IsLeaf: true, ClassLabel: maj, MajorityClass: maj, Depth: depth}
		}
		if cfg.Pre.UseK && len(ds.Data) < cfg.Pre.K {
			return &Node{IsLeaf: true, ClassLabel: maj, MajorityClass: maj, Depth: depth}
		}
	}

	if len(ds.Attrs) == 0 {
		return &Node{IsLeaf: true, ClassLabel: maj, MajorityClass: maj, Depth: depth}
	}

	bestAttr := ""
	bestIG := -1.0
	baseH := entropy(ds.Data, ds.ClassAttr)
	for _, a := range ds.Attrs {
		ig := baseH - condEntropy(ds.Data, a, ds.ClassAttr)
		if ig > bestIG {
			bestIG = ig
			bestAttr = a
		}
	}

	if bestAttr == "" {
		return &Node{IsLeaf: true, ClassLabel: maj, MajorityClass: maj, Depth: depth}
	}

	if (cfg.Mode == "0" || cfg.Mode == "2") && cfg.Pre.UseG {
		if bestIG < cfg.Pre.G {
			return &Node{IsLeaf: true, ClassLabel: maj, MajorityClass: maj, Depth: depth}
		}
	}

	children := make(map[string]*Node)
	valGroups := groupByValue(ds.Data, bestAttr)

	rem := make([]string, 0, len(ds.Attrs)-1)
	for _, a := range ds.Attrs {
		if a != bestAttr {
			rem = append(rem, a)
		}
	}

	for v, subset := range valGroups {
		subds := ds
		subds.Attrs = rem
		subds.Data = subset
		if len(subset) == 0 {
			children[v] = &Node{IsLeaf: true, ClassLabel: maj, MajorityClass: maj, Depth: depth + 1}
		} else {
			children[v] = buildID3(subds, cfg, depth+1)
		}
	}

	return &Node{
		IsLeaf:        false,
		Attr:          bestAttr,
		Children:      children,
		MajorityClass: maj,
		Depth:         depth,
	}
}

func entropy(data []Example, classAttr string) float64 {
	counts := make(map[string]int)
	for _, ex := range data {
		counts[ex[classAttr]]++
	}
	n := float64(len(data))
	h := 0.0
	for _, c := range counts {
		p := float64(c) / n
		if p > 0 {
			h -= p * log2(p)
		}
	}
	return h
}

func condEntropy(data []Example, attr string, classAttr string) float64 {
	parts := groupByValue(data, attr)
	n := float64(len(data))
	sum := 0.0
	for _, subset := range parts {
		if len(subset) == 0 {
			continue
		}
		w := float64(len(subset)) / n
		sum += w * entropy(subset, classAttr)
	}
	return sum
}

func groupByValue(data []Example, attr string) map[string][]Example {
	out := make(map[string][]Example)
	for _, ex := range data {
		v := ex[attr]
		out[v] = append(out[v], ex)
	}
	return out
}

func isPure(data []Example, classAttr string) bool {
	if len(data) == 0 {
		return true
	}
	first := data[0][classAttr]
	for _, ex := range data[1:] {
		if ex[classAttr] != first {
			return false
		}
	}
	return true
}

func majorityClass(data []Example, classAttr string) string {
	counts := make(map[string]int)
	for _, ex := range data {
		counts[ex[classAttr]]++
	}
	best := ""
	bestN := -1
	for k, v := range counts {
		if v > bestN {
			bestN = v
			best = k
		}
	}
	return best
}

func log2(x float64) float64 {
	return math.Log(x) / math.Log(2)
}

func predict(root *Node, ex Example) string {
	if root.IsLeaf {
		return root.ClassLabel
	}
	v := ex[root.Attr]
	if child, ok := root.Children[v]; ok {
		return predict(child, ex)
	}
	return root.MajorityClass
}

func accuracy(root *Node, ds Dataset) float64 {
	if len(ds.Data) == 0 {
		return 0
	}
	correct := 0
	for _, ex := range ds.Data {
		if predict(root, ex) == ex[ds.ClassAttr] {
			correct++
		}
	}
	return float64(correct) / float64(len(ds.Data))
}

func reducedErrorPrune(root *Node, val Dataset, classAttr string) *Node {
	if root == nil || root.IsLeaf {
		return root
	}

	for k, ch := range root.Children {
		root.Children[k] = reducedErrorPrune(ch, val, classAttr)
	}

	origAcc := accuracy(root, val)
	leaf := &Node{IsLeaf: true, ClassLabel: root.MajorityClass, MajorityClass: root.MajorityClass, Depth: root.Depth}
	leafAcc := accuracy(leaf, val)

	if leafAcc >= origAcc {
		return leaf
	}
	return root
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	if len(xs) == 1 {
		return mean, 0
	}
	vars := 0.0
	for _, x := range xs {
		d := x - mean
		vars += d * d
	}
	vars /= float64(len(xs) - 1)
	return mean, math.Sqrt(vars)
}

func (n *Node) String() string {
	var b strings.Builder
	printNode(&b, n, 0)
	return b.String()
}

func printNode(b *strings.Builder, n *Node, indent int) {
	pad := strings.Repeat("  ", indent)
	if n.IsLeaf {
		fmt.Fprintf(b, "%s[LEAF] %s\n", pad, n.ClassLabel)
		return
	}
	fmt.Fprintf(b, "%s[%s] (maj=%s)\n", pad, n.Attr, n.MajorityClass)
	// stable order
	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, "%s  - %s:\n", pad, k)
		printNode(b, n.Children[k], indent+2)
	}
}

func atoiSafe(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
