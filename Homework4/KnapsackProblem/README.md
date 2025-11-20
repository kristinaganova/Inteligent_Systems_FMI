# Knapsack Problem

Solve the **Knapsack Problem (KP)** using a **Genetic Algorithm (GA)**.

The goal is to select a **subset of items** with **maximum value** such that:
- the **total weight** does **not exceed the capacity** of the knapsack.

---

## Input

- One line with two integers:
  ```text
  M N
  ```
where:
•	M – knapsack capacity
•	N – number of items
•	Followed by N lines with two integers each:

  ```text
mi ci
  ```
where:
	•	mi – weight of the i-th item
	•	ci – value of the i-th item

## Output
	1.	First block – at least 10 values, one per line:
	  -	the current best value in the population:
	  -	first generation
	  -	at least eight intermediate generations
	  -	last generation
	2.	Empty line.
	3.	Final maximum value (one value), equal to the last value from the block above.

---

## Notes
	-	It is expected to reach the optimum in most runs (at least 8 out of 10).
	-	The solution should work within seconds, even for large inputs (N ≤ 10,000).
	-	Provide an option to measure and print the time to find the solution.

The first line in each file contains the maximum capacity and the number of items.

## Sample Optimal Values:
	-	"Short" dataset:
	-	capacity: 5000
	-	number of items: 24
	-	optimum: 1130
	-	"Long" dataset:
	-	capacity: 5000
	-	number of items: 200
	-	optimum: 5119

