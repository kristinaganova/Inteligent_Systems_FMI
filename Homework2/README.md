### Sliding Blocks

The game starts with a **square board** consisting of tiles numbered from `1` to `N` and one empty tile represented by the digit `0`.  
The goal is to **arrange the tiles in order according to their numbers**.

Movement is performed by moving tiles into the empty tile position from:

- above  
- below  
- left  
- right  

---

## Input

The input consists of:

1. Number `N` – the number of numbered tiles (e.g.: `8`, `15`, `24`, etc.).
2. Number `I` – the index of the zero position in the solution:  
   - if `I = -1` – use the **default index**: zero is **bottom right**;
   - otherwise – use the provided index.
3. The board arrangement – elements (numbers) provided in the form of a square matrix.

---

## Task

Using the **IDA\*** algorithm and the **"Manhattan distance"** heuristic, output:

1. On the first line – the **length of the optimal path** from the start to the goal state.
2. On the following lines – the corresponding **steps** leading to the solution:

   - `left`
   - `right`
   - `up`
   - `down`

---

## Notes

- Not every input puzzle configuration is **solvable**.  
  A solvability check can be performed using the methods described in the corresponding reference link.

- For **even boards** and a **goal state** where `0` is at the **first index**,  
  the check rule must be **modified** from the standard algorithm.

- For an **unsolvable puzzle** the output is: -1

## Sample Input: 
```text
8
-1
1 2 3
4 5 6
0 7 8
```

## Sample Output: 
```text
2
left
left
```
