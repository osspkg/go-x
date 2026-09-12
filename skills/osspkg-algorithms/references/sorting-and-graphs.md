# Sorting and topological ordering

## Sorting: `sorts`

Import:

```go
import "go.osspkg.com/algorithms/sorts"
```

All functions mutate the input slice. The `less(i, j)` function must return
`true` when the element at index `i` should come before the element at index `j`
in the current slice:

```go
type Job struct {
	Name     string
	Priority int
}

jobs := []Job{
	{Name: "backup", Priority: 30},
	{Name: "deploy", Priority: 10},
	{Name: "test", Priority: 20},
}

sorts.Merge(jobs, func(i, j int) bool {
	return jobs[i].Priority < jobs[j].Priority
})
// jobs: deploy, test, backup
```

Available functions are `Bubble`, `Cocktail`, `Heapsort`, `Insertion`, `Merge`,
`Selection`, and `Reverse`. For ordinary application sorting, first compare the
requirements with `slices.SortFunc`/`sort.Slice`; when stability is required,
explicitly use a stable standard-library sort because stability is not a general
contract of these functions.

Practical selection:

- `Merge` and `Heapsort` — for the general `O(n log n)` case;
- `Insertion` — for small or nearly sorted slices;
- `Bubble`, `Cocktail`, and `Selection` — for educational or narrow cases with
  small data sets;
- `Reverse` — a linear reversal without a callback.

Do not pass a callback that reads a separate copy of the slice: the algorithm
reorders elements while running and passes indices for the current state.

## DFS graph: `graph/dfs`

The DFS variant supports generic `comparable` keys, requires both vertices to be
added before an edge, and returns an error for a missing vertex or a cycle:

```go
package main

import (
	"errors"
	"fmt"

	"go.osspkg.com/algorithms/graph/dfs"
)

func main() {
	g := dfs.NewGraph[string]()
	for _, node := range []string{"clean", "compile", "test", "deploy"} {
		if err := g.AddNode(node); err != nil {
			panic(err)
		}
	}
	for _, edge := range [][2]string{
		{"clean", "compile"},
		{"compile", "test"},
		{"test", "deploy"},
	} {
		if err := g.AddEdge(edge[0], edge[1]); err != nil {
			panic(err)
		}
	}

	order, err := g.TopologicalSort()
	if errors.Is(err, dfs.ErrCycleDetected) {
		panic("dependency cycle")
	}
	if err != nil {
		panic(err)
	}
	fmt.Println(order)
}
```

The edge direction `from -> to` means that `from` must come before `to`. For
independent vertices, define your own stabilization rule: because the graph
iterates over a map, the complete order of such vertices is not guaranteed.

## Kahn graph: `graph/kahn`

This variant works with `string` keys, registers vertices automatically through
`Add`, sorts the initial queue by key, and exposes the result through `Result`:

```go
package main

import (
	"errors"
	"fmt"

	"go.osspkg.com/algorithms/graph/kahn"
)

func main() {
	g := kahn.New()
	g.Add("base", "lib")
	g.Add("lib", "app")
	g.Add("unrelated", "unused")
	g.BreakPoint("app")

	if err := g.Build(); err != nil {
		if errors.Is(err, kahn.ErrBreakPoint) {
			panic("break point is absent")
		}
		if errors.Is(err, kahn.ErrBuild) {
			panic("cycle in active dependencies")
		}
		panic(err)
	}
	fmt.Println(g.Result()) // [base lib app]
}
```

`BreakPoint` keeps only the target vertex and its transitive dependencies
(ancestors); unrelated vertices and cycles outside this subgraph do not affect
`Build`. Without `BreakPoint`, the complete graph is built. Edge direction is
the same: add a dependency first, for example `g.Add("base", "lib")`.

Verified examples in the checkout: `graph/dfs/graph_test.go` and
`graph/kahn/graph_test.go`.
