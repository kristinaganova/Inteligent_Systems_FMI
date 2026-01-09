import argparse
import math
import random
import ssl
from dataclasses import dataclass, field
from typing import Dict, List, Tuple

import pandas as pd
from ucimlrepo import fetch_ucirepo
from ucimlrepo.fetch import DatasetNotFoundError

# Disable SSL certificate verification so ucimlrepo can fetch the dataset
# even if local certificates are not configured (NOT for production use).
ssl._create_default_https_context = ssl._create_unverified_context

Example = Dict[str, str]


@dataclass
class Dataset:
    attrs: List[str]
    class_attr: str
    data: List[Example]
    attr_vals: Dict[str, List[str]] = field(default_factory=dict)


@dataclass
class Node:
    is_leaf: bool
    class_label: str = ""
    attr: str = ""
    children: Dict[str, "Node"] = field(default_factory=dict)
    majority_class: str = ""
    depth: int = 0


@dataclass
class PrePrune:
    use_n: bool = False
    n: int = 0
    use_k: bool = False
    k: int = 0
    use_g: bool = False
    g: float = 0.0


@dataclass
class Config:
    mode: str = "2"  # "0","1","2"
    pre: PrePrune = field(default_factory=PrePrune)
    use_post_e: bool = True
    seed: int = 42
    train_ratio: float = 0.8
    val_ratio_in_train: float = 0.2
    folds: int = 10
    missing_policy: str = "mode_by_class"


def default_config() -> Config:
    return Config(
        mode="2",
        pre=PrePrune(use_n=True, n=10, use_k=True, k=5, use_g=True, g=0.1),
        use_post_e=True,
        seed=42,
        train_ratio=0.8,
        val_ratio_in_train=0.2,
        folds=10,
        missing_policy="mode_by_class",
    )


# input examples:
# "0" => all implemented pre-pruning (N,K,G), no post
# "0 K" => only K
# "1" => all implemented post-pruning (E), no pre
# "1 E" => only E
# "2" => both all
# "2 NKG E" => both, only specified subsets

def parse_input_pruning(s: str, cfg: Config) -> None:
    parts = s.strip().split()
    if not parts:
        return
    mode = parts[0]
    cfg.mode = mode

    if mode == "0":
        cfg.pre.use_n = cfg.pre.use_k = cfg.pre.use_g = True
        cfg.use_post_e = False
    elif mode == "1":
        cfg.pre.use_n = cfg.pre.use_k = cfg.pre.use_g = False
        cfg.use_post_e = True
    elif mode == "2":
        cfg.pre.use_n = cfg.pre.use_k = cfg.pre.use_g = True
        cfg.use_post_e = True

    if len(parts) > 1:
        if mode in {"0", "2"}:
            cfg.pre.use_n = cfg.pre.use_k = cfg.pre.use_g = False
        if mode in {"1", "2"}:
            cfg.use_post_e = False

        for t in parts[1:]:
            up = t.upper()
            if any(c in up for c in "NKG") and mode in {"0", "2"}:
                if "N" in up:
                    cfg.pre.use_n = True
                if "K" in up:
                    cfg.pre.use_k = True
                if "G" in up:
                    cfg.pre.use_g = True
            if "E" in up and mode in {"1", "2"}:
                cfg.use_post_e = True

        if mode in {"0", "2"} and not (cfg.pre.use_n or cfg.pre.use_k or cfg.pre.use_g):
            cfg.pre.use_n = cfg.pre.use_k = cfg.pre.use_g = True
        if mode in {"1", "2"} and not cfg.use_post_e:
            cfg.use_post_e = True


def load_arff(path: str) -> Dataset:
    attr_vals: Dict[str, List[str]] = {}
    data: List[Example] = []
    all_attrs: List[str] = []
    in_data = False

    with open(path, "r", encoding="utf-8") as f:
        for raw_line in f:
            line = raw_line.strip()
            if not line or line.startswith("%"):
                continue
            low = line.lower()
            if low.startswith("@relation"):
                continue
            if low.startswith("@attribute"):
                toks = line.split()
                if len(toks) < 3:
                    raise ValueError(f"bad @attribute line: {line}")
                name = toks[1]
                spec = " ".join(toks[2:])
                all_attrs.append(name)
                if "{" in spec and "}" in spec:
                    vals = parse_brace_list(spec)
                    attr_vals[name] = vals
                continue
            if low.startswith("@data"):
                in_data = True
                continue
            if in_data:
                parts = split_csv_line(line)
                if len(parts) != len(all_attrs):
                    raise ValueError(
                        f"data row has {len(parts)} cols, expected {len(all_attrs)}: {line}"
                    )
                ex: Example = {}
                for i, a in enumerate(all_attrs):
                    v = parts[i].strip()
                    if v == "?":
                        v = ""
                    ex[a] = v
                data.append(ex)

    if not all_attrs or not data:
        raise ValueError("no attributes or no data found")

    class_attr = all_attrs[-1]
    attrs = all_attrs[:-1]
    return Dataset(attrs=attrs, class_attr=class_attr, data=data, attr_vals=attr_vals)


def parse_brace_list(spec: str) -> List[str]:
    l = spec.find("{")
    r = spec.rfind("}")
    if l < 0 or r < 0 or r <= l:
        return []
    inside = spec[l + 1 : r]
    raw = inside.split(",")
    return [x.strip() for x in raw]


def split_csv_line(line: str) -> List[str]:
    return line.split(",")


def load_breast_cancer_uciml() -> Dataset:
    # fetch dataset
    breast_cancer = fetch_ucirepo(id=14)

    # data (as pandas dataframes)
    X = breast_cancer.data.features
    y = breast_cancer.data.targets

    # assume single target column
    class_attr = y.columns[0]
    attrs = list(X.columns)

    data: List[Example] = []
    for i in range(len(X)):
        ex: Example = {}
        for a in attrs:
            v = X.iloc[i][a]
            if pd.isna(v):
                v = ""
            ex[a] = str(v)
        cval = y.iloc[i][class_attr]
        if pd.isna(cval):
            cval = ""
        ex[class_attr] = str(cval)
        data.append(ex)

    # optional: collect possible values per attribute (not critical for the algorithm)
    attr_vals: Dict[str, List[str]] = {}
    for a in attrs:
        vals = sorted({str(v) for v in X[a].dropna().unique()})
        attr_vals[a] = vals
    vals_class = sorted({str(v) for v in y[class_attr].dropna().unique()})
    attr_vals[class_attr] = vals_class

    # show metadata and variables like in the example
    print(breast_cancer.metadata)
    print(breast_cancer.variables)

    return Dataset(attrs=attrs, class_attr=class_attr, data=data, attr_vals=attr_vals)


def impute_missing_mode_by_class(ds: Dataset) -> None:
    class_counts: Dict[str, int] = {}
    counts: Dict[str, Dict[str, Dict[str, int]]] = {}
    global_counts: Dict[str, Dict[str, int]] = {}

    for ex in ds.data:
        c = ex[ds.class_attr]
        class_counts[c] = class_counts.get(c, 0) + 1
        if c not in counts:
            counts[c] = {}
        for a in ds.attrs:
            v = ex[a]
            if v == "":
                continue
            if a not in counts[c]:
                counts[c][a] = {}
            counts[c][a][v] = counts[c][a].get(v, 0) + 1

            if a not in global_counts:
                global_counts[a] = {}
            global_counts[a][v] = global_counts[a].get(v, 0) + 1

    mode_by_class: Dict[str, Dict[str, str]] = {}
    global_mode: Dict[str, str] = {}

    for a in ds.attrs:
        global_mode[a] = argmax(global_counts.get(a, {}))

    for c in class_counts.keys():
        mode_by_class[c] = {}
        for a in ds.attrs:
            mode = argmax(counts.get(c, {}).get(a, {}))
            if not mode:
                mode = global_mode.get(a, "")
            mode_by_class[c][a] = mode

    for ex in ds.data:
        c = ex[ds.class_attr]
        for a in ds.attrs:
            if ex[a] == "":
                ex[a] = mode_by_class[c].get(a, ex[a])


def argmax(m: Dict[str, int]) -> str:
    best = ""
    best_n = -1
    for k, v in m.items():
        if v > best_n:
            best_n = v
            best = k
    return best


def stratified_split(ds: Dataset, train_ratio: float, rng: random.Random) -> Tuple[Dataset, Dataset]:
    by_class: Dict[str, List[Example]] = {}
    for ex in ds.data:
        c = ex[ds.class_attr]
        by_class.setdefault(c, []).append(ex)

    train_data: List[Example] = []
    test_data: List[Example] = []

    for arr in by_class.values():
        shuffled = list(arr)
        rng.shuffle(shuffled)
        split = int(round(len(shuffled) * train_ratio))
        if split < 1:
            split = 1
        if split > len(shuffled) - 1:
            split = len(shuffled) - 1
        train_data.extend(shuffled[:split])
        test_data.extend(shuffled[split:])

    train = Dataset(attrs=ds.attrs, class_attr=ds.class_attr, data=train_data, attr_vals=ds.attr_vals)
    test = Dataset(attrs=ds.attrs, class_attr=ds.class_attr, data=test_data, attr_vals=ds.attr_vals)
    return train, test


def stratified_k_fold_cv(train: Dataset, k: int, cfg: Config, rng: random.Random) -> List[float]:
    folds: List[List[Example]] = [[] for _ in range(k)]
    by_class: Dict[str, List[Example]] = {}
    for ex in train.data:
        by_class.setdefault(ex[train.class_attr], []).append(ex)

    for arr in by_class.values():
        shuffled = list(arr)
        rng.shuffle(shuffled)
        for i, ex in enumerate(shuffled):
            folds[i % k].append(ex)

    accs: List[float] = []
    for i in range(k):
        tr: List[Example] = []
        te: List[Example] = []
        for j in range(k):
            if j == i:
                te.extend(folds[j])
            else:
                tr.extend(folds[j])
        ds_tr = Dataset(attrs=train.attrs, class_attr=train.class_attr, data=tr, attr_vals=train.attr_vals)
        ds_te = Dataset(attrs=train.attrs, class_attr=train.class_attr, data=te, attr_vals=train.attr_vals)

        model = train_with_optional_post_pruning(ds_tr, cfg, rng)
        accs.append(accuracy(model, ds_te))

    return accs


def train_with_optional_post_pruning(train: Dataset, cfg: Config, rng: random.Random) -> Node:
    if cfg.use_post_e:
        subtrain, val = stratified_split(train, 1.0 - cfg.val_ratio_in_train, rng)
        tree = build_id3(subtrain, cfg, 0)
        tree = reduced_error_prune(tree, val, train.class_attr)
        return tree
    return build_id3(train, cfg, 0)


def build_id3(ds: Dataset, cfg: Config, depth: int) -> Node:
    maj = majority_class(ds.data, ds.class_attr)

    if is_pure(ds.data, ds.class_attr):
        return Node(is_leaf=True, class_label=ds.data[0][ds.class_attr], majority_class=maj, depth=depth)

    if cfg.mode in {"0", "2"} and (cfg.pre.use_n or cfg.pre.use_k or cfg.pre.use_g):
        if cfg.pre.use_n and depth >= cfg.pre.n:
            return Node(is_leaf=True, class_label=maj, majority_class=maj, depth=depth)
        if cfg.pre.use_k and len(ds.data) < cfg.pre.k:
            return Node(is_leaf=True, class_label=maj, majority_class=maj, depth=depth)

    if not ds.attrs:
        return Node(is_leaf=True, class_label=maj, majority_class=maj, depth=depth)

    best_attr = ""
    best_ig = -1.0
    base_h = entropy(ds.data, ds.class_attr)
    for a in ds.attrs:
        ig = base_h - cond_entropy(ds.data, a, ds.class_attr)
        if ig > best_ig:
            best_ig = ig
            best_attr = a

    if not best_attr:
        return Node(is_leaf=True, class_label=maj, majority_class=maj, depth=depth)

    if cfg.mode in {"0", "2"} and cfg.pre.use_g:
        if best_ig < cfg.pre.g:
            return Node(is_leaf=True, class_label=maj, majority_class=maj, depth=depth)

    children: Dict[str, Node] = {}
    val_groups = group_by_value(ds.data, best_attr)

    rem = [a for a in ds.attrs if a != best_attr]

    for v, subset in val_groups.items():
        subds = Dataset(attrs=rem, class_attr=ds.class_attr, data=subset, attr_vals=ds.attr_vals)
        if not subset:
            children[v] = Node(is_leaf=True, class_label=maj, majority_class=maj, depth=depth + 1)
        else:
            children[v] = build_id3(subds, cfg, depth + 1)

    return Node(
        is_leaf=False,
        attr=best_attr,
        children=children,
        majority_class=maj,
        depth=depth,
    )


def entropy(data: List[Example], class_attr: str) -> float:
    counts: Dict[str, int] = {}
    for ex in data:
        counts[ex[class_attr]] = counts.get(ex[class_attr], 0) + 1
    n = float(len(data))
    h = 0.0
    for c in counts.values():
        p = float(c) / n
        if p > 0:
            h -= p * math.log2(p)
    return h


def cond_entropy(data: List[Example], attr: str, class_attr: str) -> float:
    parts = group_by_value(data, attr)
    n = float(len(data))
    s = 0.0
    for subset in parts.values():
        if not subset:
            continue
        w = float(len(subset)) / n
        s += w * entropy(subset, class_attr)
    return s


def group_by_value(data: List[Example], attr: str) -> Dict[str, List[Example]]:
    out: Dict[str, List[Example]] = {}
    for ex in data:
        v = ex[attr]
        out.setdefault(v, []).append(ex)
    return out


def is_pure(data: List[Example], class_attr: str) -> bool:
    if not data:
        return True
    first = data[0][class_attr]
    for ex in data[1:]:
        if ex[class_attr] != first:
            return False
    return True


def majority_class(data: List[Example], class_attr: str) -> str:
    counts: Dict[str, int] = {}
    for ex in data:
        counts[ex[class_attr]] = counts.get(ex[class_attr], 0) + 1
    best = ""
    best_n = -1
    for k, v in counts.items():
        if v > best_n:
            best_n = v
            best = k
    return best


def predict(root: Node, ex: Example) -> str:
    if root.is_leaf:
        return root.class_label
    v = ex.get(root.attr, "")
    child = root.children.get(v)
    if child is not None:
        return predict(child, ex)
    return root.majority_class


def accuracy(root: Node, ds: Dataset) -> float:
    if not ds.data:
        return 0.0
    correct = 0
    for ex in ds.data:
        if predict(root, ex) == ex[ds.class_attr]:
            correct += 1
    return correct / float(len(ds.data))


def reduced_error_prune(root: Node, val: Dataset, class_attr: str) -> Node:
    if root is None or root.is_leaf:
        return root

    for k, ch in list(root.children.items()):
        root.children[k] = reduced_error_prune(ch, val, class_attr)

    orig_acc = accuracy(root, val)
    leaf = Node(is_leaf=True, class_label=root.majority_class, majority_class=root.majority_class, depth=root.depth)
    leaf_acc = accuracy(leaf, val)

    if leaf_acc >= orig_acc:
        return leaf
    return root


def mean_std(xs: List[float]) -> Tuple[float, float]:
    if not xs:
        return 0.0, 0.0
    s = sum(xs)
    mean = s / float(len(xs))
    if len(xs) == 1:
        return mean, 0.0
    var = sum((x - mean) ** 2 for x in xs) / float(len(xs) - 1)
    return mean, math.sqrt(var)


def node_to_string(n: Node, indent: int = 0) -> str:
    pad = "  " * indent
    if n.is_leaf:
        return f"{pad}[LEAF] {n.class_label}\n"
    lines = [f"{pad}[{n.attr}] (maj={n.majority_class})\n"]
    for k in sorted(n.children.keys()):
        lines.append(f"{pad}  - {k}:\n")
        lines.append(node_to_string(n.children[k], indent + 2))
    return "".join(lines)


def main() -> None:
    parser = argparse.ArgumentParser(description="ID3 decision tree with pruning (Python version)")
    parser.add_argument(
        "--input",
        default="2",
        help='pruning input like: "0", "0 K", "1 E", "2 NKG E"',
    )
    parser.add_argument("--seed", type=int, default=42, help="random seed")
    args = parser.parse_args()

    cfg = default_config()
    cfg.seed = args.seed
    parse_input_pruning(args.input, cfg)

    # load Breast Cancer dataset via ucimlrepo
    try:

        ds = load_breast_cancer_uciml()
    except DatasetNotFoundError as e:
        print("Error: could not fetch Breast Cancer dataset via ucimlrepo (id=14).")
        print("Most common cause: SSL / certificate problem on this machine.")
        print("Details:", e)
        return
    except Exception as e:
        print("Unexpected error while fetching dataset via ucimlrepo:")
        print(e)
        return

    impute_missing_mode_by_class(ds)

    rng = random.Random(cfg.seed)
    train, test = stratified_split(ds, cfg.train_ratio, rng)

    model = train_with_optional_post_pruning(train, cfg, rng)
    train_acc = accuracy(model, train)

    fold_accs = stratified_k_fold_cv(train, cfg.folds, cfg, rng)
    avg, std = mean_std(fold_accs)

    model2 = train_with_optional_post_pruning(train, cfg, rng)
    test_acc = accuracy(model2, test)

    print(f"Train set accuracy: {train_acc:.2%}")

    print(f"------ Performing {cfg.folds}-Fold Cross-Validation ------")
    for i, a in enumerate(fold_accs):
        print(f"[FOLD {i}] Accuracy: {a:.2%}")
    print(f"Average acuracy: {avg:.2%}")
    print(f"Standard deviation: {std:.2%}")
    print(f"------ Validation completed ------")

    print(f"Test set accuracy: {test_acc:.2%}")


if __name__ == "__main__":
    main()
