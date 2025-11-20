### N-Queens

The **N-Queens** problem consists of placing `N` queens on a square `N x N` chessboard such that they **do not attack each other** – i.e., there is no more than one queen in the same:

- row  
- column  
- diagonal  

Use the **Min-Conflicts** algorithm to find a solution.

---

## Input

- An integer `N` – the number of queens (`N ≥ 1`).

---

## Output

An array with the positions of the queens, where:

- the **index** corresponds to the **column**
- the **value** corresponds to the **row**

Example:

For `N = 4` a possible output is:

```text
[1, 3, 0, 2]
```

## Notes
-	The problem has no solution for N ∈ {2, 3}.
- In these cases the output should be: -1
- Provide an option to measure and print the time to find the solution
(more details are available in the "Automated Testing" section).
- For large values of N it is not necessary to output the solution.
In these cases only the execution time of the algorithm can be printed
(more details are available in the "Automated Testing" section).
	•	Support an option for printing the board (for example via a flag or hardcoded setting):
	•	Queen: *
	•	Empty cell: _
Each row is printed on a new line, cells are separated by spaces.
Example output for N = 4:

```text
_ * _ _
_ _ _ *
* _ _ _
_ _ * _
```
