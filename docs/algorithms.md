# Algorithms

## Max Probability Path (Modified Dijkstra)

Finds the single path with the highest joint probability. The joint probability of a path is the product of its edge probabilities.

The algorithm converts this to a shortest-path problem using the identity:

```
max product(P_i) = min sum(-log(P_i))
```

This allows standard Dijkstra's algorithm to be applied with `-log(probability)` as edge weights.

**Complexity:** O((V + E) log V)

## Top-K Probability Paths (Yen's Algorithm)

Extends the max probability path to find the K most probable paths. Based on Yen's algorithm for K-shortest paths:

1. Find the best path using modified Dijkstra
2. For each previously found path, systematically generate candidate "spur" paths by removing edges that would duplicate known paths
3. Select the best candidate and repeat until K paths are found

## Exact Reachability (DFS with Memoization)

Computes the exact probability that a target is reachable from a source across all possible edge activation combinations:

```
P(s -> t) = 1 - product over children c of (1 - P(edge_sc) * P(c -> t))
```

Uses DFS with memoization to avoid recomputing sub-problems. Handles cycles by tracking visited nodes.

**Complexity:** O(V + E)

## Monte Carlo Reachability

Estimates reachability by sampling:

1. For each sample, independently activate each edge with its probability (Bernoulli trial)
2. Test reachability in the resulting deterministic graph using BFS
3. The fraction of samples where the target is reachable estimates the true probability
4. Returns a 95% confidence interval: `estimate ± 1.96 * sqrt(p(1-p) / n)`

Uses parallel workers (one per CPU core), each with an independent PCG random number generator. Default: 10,000 samples.
