# Tic-Tac-Toe

The task is to implement a **deterministic agent** for playing **Tic-Tac-Toe** that plays **optimally**, using the **Minimax algorithm with α–β pruning**.

The program supports two operating modes:

- `JUDGE` – used by the testing tool (**required**).
- `GAME` – interactive human–computer game (for local experiments and demonstrations, **also required**).

All boards are printed in **"framed" format** (3×3), for example:

```text
+---+---+---+
| X | O | _ |
+---+---+---+
| _ | X | _ |
+---+---+---+
| _ | _ | O |
+---+---+---+
```

The symbol `_` represents an empty cell.

The first line of standard input always indicates the mode: **JUDGE** or **GAME**.

---

## JUDGE Mode

### Input

...
JUDGE  
TURN X        # or TURN O  
<3×3 board in framed format – exactly 7 lines>  
...

### Output (accepted by the testing tool)

- For **non-terminal positions**: the chosen move — two integers (1-based):

...
row col  
...

where `row ∈ {1,2,3}`, `col ∈ {1,2,3}`.

- For **terminal positions**:

...
-1  
...

### Sample Input

```text
JUDGE  
TURN X  
+---+---+---+  
| _ | _ | _ |  
+---+---+---+  
| _ | _ | _ |  
+---+---+---+  
| _ | _ | _ |  
+---+---+---+  
```

### Sample Output

```text
2 2  
```

---

## GAME Mode

### Input

```text
GAME  
FIRST X       # who starts  
HUMAN O       # which side is the human  
<3×3 initial board position>  
```

### Program Requirements

1. Read the initial board.
2. Determine who is human and who is the agent according to `FIRST` and `HUMAN`.
3. While the game is not terminal:
   - **If it's the human's turn**:
     - read two integers `row col`
     - apply the move if valid
   - **If it's the agent's turn**:
     - find the optimal move using **Minimax with α–β pruning**
     - apply the move
   - After each move, print the updated board:
     ```text
     +---+---+---+
     | ... | ... | ... |
     +---+---+---+
     | ... | ... | ... |
     +---+---+---+
     | ... | ... | ... |
     +---+---+---+
     ```
4. In a terminal state, print one line:
   - `WINNER: X`
   - `WINNER: O`
   - `DRAW`

---
