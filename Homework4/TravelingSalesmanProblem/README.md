# Traveling Salesman Problem

Solve the **Traveling Salesman Problem (TSP)** using a **Genetic Algorithm (GA)**.

The goal is to find the **shortest possible route** that:
- passes through **all points exactly once**;
- a **path exists** between every pair of points;
- the starting point **does not matter** – it is **not necessary** for the route to return to the starting position.

---

## Input

One of the following:

1. **Random set of points**  
   - An integer `N` (`N ≤ 100`)  
   - Generate `N` **random points in the plane**.

2. **Named dataset**  
   - A string (dataset name) containing letters (and possibly digits)  
   - Followed by the number of cities  
   - Followed by lines of the form:
     ```text
     <city_name> X Y
     ```

---

## Output

1. **First block** – at least **10 values**, **one per line**:  
   - the current **best path length** in the population:
     - first generation
     - at least eight intermediate generations
     - last generation

2. **Empty line**.

3. **Best found solution**:

   - For **N random points**:  
     - only the **final length** (one value),  
       equal to the **last** value from the block above.

   - For a **named dataset**:
     - one line with the route, for example:  
       ```text
       CityA -> CityB -> ... -> CityZ
       ```
     - followed by a line with the **final length**.

> The final length must:
> - match the **last** value from the first block;
> - match the **recalculated length** of the route.

---

## Notes

- For case 2 (named dataset) it is expected to **reach the optimum** in most runs  
  (at least **8 out of 10**), with small tolerance for **floating point**.

- The solution should work **within seconds**, even for larger inputs.

- Provide an option to **measure and print the time** to find the solution  
  (more details are available in the "Automated Testing" section).

- For a sample set of cities you can use the files:
  - `uk12_name.csv`
  - `uk12_xy.csv`  
  from the `UK_TSP.zip` archive.

- The optimal solution for the **UK12** dataset has length:
  ```text
  1595.738522033024
  ```
  ## Sample Input
  ```text
  UK12
  12
  Aberystwyth 0.190032E-03 -0.285946E-03
  ...
  Stratford 217.343 -447.089
  ```
  ## Sample Output
  ```text
  2426.8086
  ...
  1595.7385
  
  Aberystwyth -> Inverness -> Nottingham -> Glasgow -> Edinburgh -> London -> Stratford -> Exeter -> Liverpool -> Oxford -> Brighton -> Newcastle
  1595.7385
  ```
  
