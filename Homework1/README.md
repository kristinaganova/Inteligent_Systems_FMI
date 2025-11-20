# Frog Puzzle

The game has a playing field of `2N + 1` squares.  
Initially:

- in the **rightmost N squares** we have frogs placed facing **left** (`<`)
- in the **leftmost N squares** we have frogs placed facing **right** (`>`)

The goal of the game is for the frogs to **swap places** and reach the **opposite configuration**.

---

## Game Rules

- Each frog can only move **in the direction it is facing**.
- Each frog can:
  - **jump to an empty space** (`_`) in front of it, or
  - **jump over one frog** (regardless of which direction it is facing) to land on an **empty space** in front of it.

---

## Task

Use **Depth-First Search (DFS)** to implement a program that solves the puzzle.

### Input

- An integer `N` – the number of frogs facing in one direction.

### Output

- All configurations that are traversed to go **from the initial to the final state** (i.e. **the steps to solve the puzzle**), each on a separate line.

---

## Constraints and Notes

- The task is expected to work with input `N = 20` in **under 1 second**.
- The task can also be solved in **linear time** using a **rule-based** approach (without full search).  
  You may try to solve it this way as an alternative solution.

---

## Example

### Sample Input
```text
2
```
### Sample Output
```text
>>_<<
>_><<
><>_<
><><_
><_<>
_<><>
<_><>
<<>_>
<<_>>
```
