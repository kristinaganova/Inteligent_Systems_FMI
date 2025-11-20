# Intelligent Systems - FMI Course

This repository contains homework assignments for the Intelligent Systems course at FMI (Faculty of Mathematics and Informatics). Each homework implements various AI algorithms and problem-solving techniques.

## Repository Structure

### Homework 1: Frog Puzzle
**Algorithm:** Depth-First Search (DFS)

A puzzle game where frogs on opposite sides of a board must swap positions. Frogs can only move in the direction they're facing and can either jump to an empty space or leap over another frog.

- **Implementation:** Go
- **Key Concepts:** State space search, DFS
- [View Details](Homework1/README.md)

---

### Homework 2: Sliding Blocks Puzzle
**Algorithm:** IDA* (Iterative Deepening A*) with Manhattan Distance Heuristic

The classic sliding puzzle where numbered tiles must be arranged in order by sliding them into the empty space.

- **Implementation:** Go
- **Key Concepts:** Informed search, heuristic functions, optimal pathfinding
- **Supported Puzzle Sizes:** 8-puzzle, 15-puzzle, 24-puzzle, etc.
- [View Details](Homework2/README.md)

---

### Homework 3: N-Queens Problem
**Algorithm:** Min-Conflicts

Place N queens on an N×N chessboard such that no two queens attack each other (no two queens share the same row, column, or diagonal).

- **Implementation:** Go
- **Key Concepts:** Constraint satisfaction, local search, conflict minimization
- **Note:** No solution exists for N ∈ {2, 3}
- [View Details](Homework3/README.md)

---

### Homework 4: Optimization Problems with Genetic Algorithms

#### 4.1 Knapsack Problem
**Algorithm:** Genetic Algorithm (GA)

Select a subset of items with maximum total value without exceeding the knapsack's weight capacity.

- **Implementation:** Go
- **Key Concepts:** Evolutionary algorithms, fitness functions, genetic operators
- **Test Cases:** 
  - Short dataset: 24 items, capacity 5000, optimum 1130
  - Long dataset: 200 items, capacity 5000, optimum 5119
- [View Details](Homework4/KnapsackProblem/README.md)

#### 4.2 Traveling Salesman Problem (TSP)
**Algorithm:** Genetic Algorithm (GA)

Find the shortest route that visits all cities exactly once.

- **Implementation:** Go
- **Key Concepts:** Evolutionary algorithms, crossover operators, mutation
- **Test Dataset:** UK12 (12 UK cities, optimal length: 1595.738522033024)
- **Scalability:** Supports random point generation up to N ≤ 100
- [View Details](Homework4/TravelingSalesmanProblem/README.md)

---

### Homework 5: Tic-Tac-Toe
**Algorithm:** Minimax with Alpha-Beta Pruning

Implement an optimal AI agent for playing Tic-Tac-Toe that never loses.

- **Implementation:** Go
- **Key Concepts:** Game theory, adversarial search, pruning techniques
- **Modes:** 
  - JUDGE mode: For automated testing
  - GAME mode: Interactive human vs. computer gameplay
- [View Details](Homework5/README.md)

---

## Technologies

- **Language:** Go (Golang)
- **Paradigm:** Procedural and functional programming
- **Focus:** Algorithm implementation and optimization

## Key Learning Outcomes

This course covers fundamental AI techniques:

1. **Search Algorithms:** DFS, IDA*, informed search
2. **Heuristic Methods:** Manhattan distance, custom heuristics
3. **Constraint Satisfaction:** Min-conflicts algorithm
4. **Evolutionary Algorithms:** Genetic algorithms, selection, crossover, mutation
5. **Game Theory:** Minimax, alpha-beta pruning
6. **Optimization:** Local search, global optimization

## Running the Programs

Each homework directory contains executable files that can be run directly. Most programs read from standard input and write to standard output.


## Performance Requirements

- Solutions are optimized for performance and meet specified time constraints
- Most problems handle large inputs efficiently (e.g., N=20 for Frog Puzzle < 1 second)
- Genetic algorithms reach optimal or near-optimal solutions in most runs (≥80% success rate)

## Author

---


