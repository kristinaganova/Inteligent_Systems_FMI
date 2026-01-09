## Homework 8 – Decision Tree (ID3) in Python

Implementation of a decision tree (ID3) on the **Breast Cancer** dataset (UCI), written in Python and using `ucimlrepo` to download the data.

### 1. Dependencies and setup

- **Python**: 3.10+ (tested with 3.14)
- **Libraries**:
  - `ucimlrepo` – to load the `Breast Cancer` dataset
  - `pandas`

Install:

```bash
pip install ucimlrepo pandas
# or, if needed
python3 -m pip install ucimlrepo pandas
```

The main code is in `main.py`.

### 2. Data loading

The code directly uses the UCI **Breast Cancer** dataset (id=14) via `ucimlrepo`:

```python
from ucimlrepo import fetch_ucirepo

breast_cancer = fetch_ucirepo(id=14)
X = breast_cancer.data.features
y = breast_cancer.data.targets
```

Then the data is converted to the internal `Dataset` format (categorical attributes + target attribute `Class`). The dataset **metadata** and **variables** are printed to the console for reference.

### 3. Handling missing values

In the original file missing values are denoted by `?`, and via `ucimlrepo` they arrive as `NaN`. In the code they are handled as follows:

- First, missing values are converted to the empty string `""`.
- Then the function `impute_missing_mode_by_class`:
  - for each class and each attribute computes the **mode per class**;
  - if a given class has no observations for an attribute, it falls back to the **global mode** of that attribute;
  - all empty values are replaced with the corresponding modal value.

This approach is suitable for **purely categorical** attributes and preserves the per-class distributions without introducing an extra artificial category.

### 4. Pre- and post-pruning

The implementation includes several techniques to avoid overfitting.

#### 4.1. Pre-pruning

The `PrePrune` configuration supports three constants:

- **N** – maximum depth of the tree.
- **K** – minimum number of training examples required to further split a node.
- **G** – minimum information gain required to perform a split on an attribute.

These conditions are checked in `build_id3` before creating new child nodes. If a pre-pruning condition is triggered, the node becomes a **leaf** labeled with the majority class.

#### 4.2. Post-pruning (Reduced Error Pruning – E)

If post-pruning is enabled, the training set is further split (stratified) into:

- **subtrain** – used to train the initial tree;
- **validation** – used to evaluate pruning decisions.

The `reduced_error_prune` algorithm:

- recursively traverses internal nodes;
- for each node compares the accuracy on the `validation` set of
  - the original subtree, and
  - a leaf with the node’s majority class;
- if the leaf does not worsen the accuracy, the subtree is replaced by the leaf.

### 5. Pruning modes (`--input`)

The program takes an `--input` argument that controls which pruning methods are used.

- **0** – use only pre-pruning (N, K, G).
- **1** – use only post-pruning (E – Reduced Error Pruning).
- **2** – use both pre- and post-pruning.

Subsets of methods can be specified with letters:

- Pre-pruning: **N**, **K**, **G**.
- Post-pruning: **E**.

Examples:

- `"0"` – all implemented pre-pruning methods (N, K, G).
- `"0 K"` – only K (minimum number of examples).
- `"1"` – all implemented post-pruning methods (E).
- `"2"` – all pre- and post-pruning methods.
- `"2 NKG E"` – N, K, G + E.

Parsing and activation/deactivation of the methods is implemented in `parse_input_pruning`.

### 6. Data splitting and cross-validation

- The data is split into **train** and **test** in an **80:20** ratio using `stratified_split`.
- The split is **stratified by class** (the original 201/85 class ratio is preserved in both subsets).
- For model evaluation a **10-fold stratified cross-validation** is performed on the training set via `stratified_k_fold_cv`.

### 7. Running the code

From the repository root:

```bash
cd Homework8
python3 main.py --input "2" --seed 42
```

Arguments:

- **`--input`** – pruning mode (see section 5).
- **`--seed`** – RNG seed for reproducibility.

### 8. Output format

Example output:

```text
Train set accuracy: 70.31%
------ Performing 10-Fold Cross-Validation ------
[FOLD 0] Accuracy: 70.83%
...
Average acuracy: 70.32%
Standard deviation: 1.33%
------ Validation completed ------
Test set accuracy: 70.18%
```

The program prints:

- **Accuracy on the training set** (model trained and evaluated on train).
- **Average accuracy and standard deviation** for 10-fold cross-validation on the training set.
- **Accuracy on the test set** (20% hold‑out, unseen during training).

### 9. Notes on results

- Obtained test accuracies are typically in the range **≈66–72%**, depending on the chosen N, K, G and whether post-pruning is used.
- Large N (deep trees) can give very high training accuracy (≈98%) but lower CV/test accuracy → clear **overfitting**.
- Restricting by **G (minimum information gain)** gives the most balanced behavior (train ≈ CV ≈ test, small standard deviation).
- Adding **Reduced Error Pruning (E)** further stabilizes the model without a noticeable loss in accuracy.
