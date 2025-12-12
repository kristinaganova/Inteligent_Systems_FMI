# k-Nearest Neighbors (kNN) Implementation

## Task

Implement the k-Nearest Neighbors (kNN) algorithm from scratch and apply it to the Iris dataset. The implementation should include:

- Data normalization (Min-Max)
- Stratified train-test split (80/20)
- kNN algorithm implementation
- kd-tree implementation for efficient nearest neighbor search
- 10-fold cross-validation
- Accuracy calculation on training and test sets

All algorithms must be implemented from scratch (no ML libraries allowed, only basic data structures).

## Solution

The program implements kNN with kd-tree optimization:

1. **Data Loading**: Reads Iris dataset from CSV file
2. **Normalization**: Applies Min-Max normalization to scale features to [0, 1]
3. **Data Splitting**: Stratified split preserving class distribution (80% train, 20% test)
4. **kd-tree**: Built from training data for efficient nearest neighbor search
5. **kNN Classification**: Predicts class using k nearest neighbors with majority voting
6. **Evaluation**: 
   - Training set accuracy
   - 10-fold cross-validation (average accuracy and standard deviation)
   - Test set accuracy

## What I Used

### Libraries (Standard Go only):
- `encoding/csv` - CSV file reading
- `math` - Mathematical operations (sqrt, abs)
- `math/rand` - Random number generation (with fixed seed for reproducibility)
- `sort` - Sorting operations
- `fmt`, `bufio`, `os` - Standard I/O

### Data Structures Implemented:
- **Point**: Stores features and class label
- **Dataset**: Collection of points
- **KDTree**: k-dimensional tree for efficient nearest neighbor search
- **Neighbor**: Stores point and distance for kNN queries

### Algorithms Implemented:
- **Min-Max Normalization**: `x_norm = (x - min) / (max - min)`
- **Stratified Splitting**: Preserves class distribution in splits
- **Euclidean Distance**: `√(Σ(x₁ᵢ - x₂ᵢ)²)`
- **kd-tree Construction**: Recursive median-based splitting
- **kNN Search**: Tree traversal with pruning optimization
- **10-Fold Cross-Validation**: Stratified folds with accuracy calculation

## Usage

```bash
# Compile
go build -o knn main.go

# Run
./knn

# Enter k value when prompted
Enter value for k: 11
```

## Output Format

```
1. Train Set Accuracy:
   Accuracy: 96.67%

2. 10-Fold Cross-Validation Results:
    Accuracy Fold 1: 100.00%
    ...
    Average Accuracy: 95.00%
    Standard Deviation: 5.53%

3. Test Set Accuracy:
   Accuracy: 100.00%
```
