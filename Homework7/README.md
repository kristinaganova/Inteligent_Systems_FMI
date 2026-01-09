# Homework 7 – Naive Bayes Classifier on Congressional Voting Records

## Task Description

Implement a **Naive Bayes Classifier** that classifies members of the U.S. House of Representatives as **democrats** or **republicans**, using the **16 voting attributes** from the
[Congressional Voting Records](https://archive.ics.uci.edu/dataset/105/congressional+voting+records) dataset.

The original data uses three symbolic values per attribute:
- `y` – voted "yes"
- `n` – voted "no"
- `?` – neither yes nor no

Although `?` usually denotes a missing value, in this dataset it explicitly means **abstained** (neither yes nor no). The assignment requires solving the problem in **two different ways**:

1. **Treat `?` as a third categorical value** ("abstained").
2. **Impute `?` with another value** using a strategy of your choice, and justify it.

Additionally, the Naive Bayes classifier may generate **zero probabilities**, which can lead to incorrect classifications. To address this, you must apply:

- **Laplace smoothing** with different values of the smoothing parameter \(\lambda\).
- **Logarithms of probabilities** (to avoid numerical underflow and products of many small numbers).

The performance of the classifier must be evaluated using:

- A single **train/test split** (80% / 20%, stratified and shuffled).
- **10-fold cross-validation** on the training set.

## Dataset

This solution downloads the data directly from the UCI repository at runtime:

- URL: `https://archive.ics.uci.edu/ml/machine-learning-databases/voting-records/house-votes-84.data`
- Format: comma-separated, first column is the **party label** (`democrat` or `republican`), followed by **16 attributes**.

The code parses all records with exactly 17 fields and normalizes empty values to `?`. Labels are stored in `y`, and the 16 attributes in `X`.

## Handling of `?` Values

The program supports **two modes**, controlled by a single integer input from `stdin`:

- **Input `0`** – **Three-valued attributes:**
  - `?` is kept as a **third valid category** ("abstained").
  - All three values (`y`, `n`, `?`) are treated symmetrically as possible feature values in the Naive Bayes model.

- **Input `1`** – **Imputed attributes:**
  - The training set is used to compute, for each attribute (column), its **most frequent non-`?` value** (column-wise mode).
  - All occurrences of `?` in both **training** and **test** sets are then replaced with this mode for that attribute.
  - This strategy is chosen because it is simple, stable for categorical data, and reflects the most common observed behavior for each vote, avoiding the introduction of unrealistic combinations.

## Model and Training Details

The implementation uses a **categorical Naive Bayes classifier** with the following characteristics:

- **Classes**: `democrat`, `republican`.
- **Features**: 16 categorical attributes (voting outcomes).
- **Class prior** \(P(c)\): estimated from relative class frequencies in the training data.
- **Conditional probabilities** \(P(x_j = v \mid c)\): estimated with **Laplace smoothing**.
- **Log probabilities**: all probabilities are stored and accumulated in log-space to avoid underflow.

### Laplace Smoothing

For each class \(c\), feature \(j\) and possible value \(v\):

- Let \(N_c\) be the number of training samples in class \(c\).
- Let \(V_j\) be the number of distinct values observed for feature \(j\) (its vocabulary size).
- Let \(\text{count}_{c,j}(v)\) be the number of times value \(v\) appears for feature \(j\) in class \(c\).

Then the smoothed probability is:

\[
P(x_j = v \mid c) = \frac{\text{count}_{c,j}(v) + \lambda}{N_c + \lambda \cdot |V_j|}
\]

The code tests multiple \(\lambda\) values:

- \(\lambda \in \{0.0, 0.1, 0.5, 1.0, 2.0\}\)

If a value is **unseen** for a particular feature in class \(c\), its probability is set to the shared **"unseen value"** probability:

\[
P(\text{unseen} \mid c) = \frac{\lambda}{N_c + \lambda \cdot |V_j|}
\]

All of these probabilities are stored as **logarithms** and combined additively when computing class scores.

### Prediction

For a sample \(x\), the classifier chooses the class with maximum posterior probability, implemented in log-space:

\[
\hat{c} = \arg\max_c \Big( \log P(c) + \sum_{j=1}^{16} \log P(x_j \mid c) \Big)
\]

## Data Splitting and Evaluation

### Stratified Train/Test Split (80% / 20%)

- The dataset is **stratified by class** so that the original proportion of democrats (267) and republicans (168) is approximately preserved in both training and test sets.
- A pseudo-random shuffle (with a time-based seed) is applied within each class before splitting.
- After stratified splitting, the training and test sets are shuffled again.

### 10-Fold Stratified Cross-Validation

- On the **training set only**, a **stratified 10-fold split** is generated.
- For each fold:
  - One fold is used as the **validation** set.
  - The remaining nine folds are used as the **training** set.
  - A fresh Naive Bayes model is trained and evaluated on this validation fold.
- The program then computes:
  - **Average accuracy** across the 10 folds.
  - **Sample standard deviation** of these accuracies.

### Model Selection and Reported Metrics

For each candidate value of \(\lambda\):

1. Train a Naive Bayes model on the **full training set**.
2. Compute **Train Set Accuracy** (on the training set).
3. Run **10-fold cross-validation** on the same training data and collect fold accuracies.
4. Compute **Average Accuracy** and **Standard Deviation** over the 10 folds.
5. Evaluate the trained model on the **held-out test set** and compute **Test Accuracy**.

All these statistics are stored, and the best \(\lambda\) is selected primarily by **highest cross-validation mean accuracy**, breaking ties with **higher test accuracy**. The final output reports:

- The **best \(\lambda\)** chosen by cross-validation.
- Detailed statistics for that model.
- A consolidated summary over all tested \(\lambda\) values.

## Input and Output Format

### Input

The program expects a single line from **standard input**:

- `0` – treat `?` as a **third category** ("abstained").
- `1` – **impute** `?` values with the **column-wise mode** (most frequent non-`?` value) computed from the training set.

Any other input value results in an error message.

### Output

The program prints:

1. A confirmation of the chosen mode.
2. The best \(\lambda\) selected by cross-validation.
3. **Train Set Accuracy** for that \(\lambda\).
4. Full **10-fold cross-validation results** (per-fold accuracies, average, and standard deviation) on the training set.
5. **Test Set Accuracy** on the held-out 20% test set.
6. A summary over all tested \(\lambda\) values (train / CV mean ± std / test accuracies).

The format is similar to:

```text
Chosen input mode: 0

===== BEST λ selected by CV mean: λ = 1.00 =====

1. Train Set Accuracy:
    Accuracy: 92.80%

10-Fold Cross-Validation Results:

    Accuracy Fold 1: 92.00%
    Accuracy Fold 2: 91.50%
    ...

    Average Accuracy: 92.10%
    Standard Deviation: 1.10%

2. Test Set Accuracy:
    Accuracy: 92.50%

----- Summary over λ (Train / CV mean±std / Test) -----
λ=0.0  | Train=92.50% | CV=91.80% ± 1.20% | Test=92.00%
λ=0.1  | Train=92.60% | CV=92.00% ± 1.10% | Test=92.40%
...
```

## How to Build and Run

### Requirements

- **Go** (version 1.20+ recommended)
- Internet access at runtime (the dataset is downloaded from the UCI repository).

### Running Directly

From the `Homework7` directory:

```bash
go run main.go
```

Then provide input via standard input, for example:

```bash
echo 0 | go run main.go
```

or

```bash
echo 1 | go run main.go
```

### Building a Binary

```bash
go build -o nb_voting main.go
./nb_voting
```

Then type `0` or `1` followed by Enter.

## Notes on Expected Results and Analysis

For a correctly implemented Naive Bayes classifier on this dataset, the **accuracy** typically falls in the range **88%–95%**, depending on:

- Whether `?` is treated as a third category or imputed.
- The chosen **smoothing parameter** \(\lambda\).
- Randomness in the train/test split and cross-validation folds.

To analyze the results:

- Compare the metrics (train accuracy, CV mean/standard deviation, and test accuracy) for:
  - **Mode 0** (three-valued attributes with `?` kept as "abstained").
  - **Mode 1** (imputed attributes).
- Observe how different values of \(\lambda\) affect overfitting and generalization:
  - Very small \(\lambda\) (including 0.0) may lead to higher variance and sensitivity to rare patterns.
  - Moderate \(\lambda\) (e.g. 0.5–1.0) often yields more stable and robust performance.

The baseline performance numbers in the UCI documentation for models such as Logistic Regression, SVM, Random Forest, Neural Networks, and XGBoost can be used as a **rough reference** for the magnitude of the accuracy, but not as strict targets, since those models differ from Naive Bayes in both modeling assumptions and capacity.
