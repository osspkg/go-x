# Algorithms

[![Go Reference](https://pkg.go.dev/badge/go.osspkg.com/algorithms.svg)](https://pkg.go.dev/go.osspkg.com/algorithms)
[![Go 1.24+](https://img.shields.io/badge/go-1.24%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)

A collection of generic algorithms, encoders, graphs, filters, and data
structures for Go.

The module is intentionally split into focused subpackages. Import only the
package you need; the module root does not expose a general-purpose API.

## Requirements

- Go 1.24 or newer

## Installation

```bash
go get go.osspkg.com/algorithms@latest
```

## Packages

| Package | Provides |
| --- | --- |
| [`comp`](https://pkg.go.dev/go.osspkg.com/algorithms/comp) | Hamming, Levenshtein, Damerau-Levenshtein, LCS, and Jaro-Winkler sequence comparison |
| [`sorts`](https://pkg.go.dev/go.osspkg.com/algorithms/sorts) | Bubble, cocktail, heap, insertion, merge, and selection sort; slice reversal |
| [`graph/dfs`](https://pkg.go.dev/go.osspkg.com/algorithms/graph/dfs) | Generic depth-first topological sorting with cycle detection |
| [`graph/kahn`](https://pkg.go.dev/go.osspkg.com/algorithms/graph/kahn) | String-keyed Kahn topological sorting with dependency subgraphs |
| [`encoding/base62`](https://pkg.go.dev/go.osspkg.com/algorithms/encoding/base62) | Configurable Base62 encoding for `uint64` values |
| [`encoding/otp`](https://pkg.go.dev/go.osspkg.com/algorithms/encoding/otp) | TOTP/HOTP generation and `otpauth` URLs |
| [`encoding/token`](https://pkg.go.dev/go.osspkg.com/algorithms/encoding/token) | Compact `T64` identifiers with text, binary, and SQL integration |
| [`structs/bitmap`](https://pkg.go.dev/go.osspkg.com/algorithms/structs/bitmap) | Concurrent growable bitmap |
| [`structs/bloom`](https://pkg.go.dev/go.osspkg.com/algorithms/structs/bloom) | Space-efficient Bloom filter with persistence support |
| [`trees/btree`](https://pkg.go.dev/go.osspkg.com/algorithms/trees/btree) | Concurrent generic B-tree |

## Usage

### Compare sequences

Comparers work with any `comparable` element type. `Similarity` returns a score
from `0` to `100`; use `[]rune` for Unicode-aware text comparison.

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/comp"
)

func main() {
	cmp := comp.NewLevenshtein[rune]()
	score := cmp.Similarity([]rune("kitten"), []rune("sitting"))

	fmt.Printf("similarity: %.2f\n", score)
}
```

Choose the comparer according to the type of difference you need to model:

- `NewHamming` compares equal-length sequences and counts positional changes.
- `NewLevenshtein` accounts for insertions, deletions, and substitutions.
- `NewDamerauLevenshtein` also accounts for adjacent transpositions.
- `NewLCS` measures the longest common subsequence.
- `NewJaroWinkler` is suited to short values and near matches.

### Sort a slice

Sorting functions mutate the supplied slice in place. The comparison callback
receives indexes into the slice's current state.

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/sorts"
)

type job struct {
	name     string
	priority int
}

func main() {
	jobs := []job{
		{name: "backup", priority: 30},
		{name: "deploy", priority: 10},
		{name: "test", priority: 20},
	}

	sorts.Merge(jobs, func(i, j int) bool {
		return jobs[i].priority < jobs[j].priority
	})

	fmt.Println(jobs) // [{deploy 10} {test 20} {backup 30}]
}
```

Available functions are `Bubble`, `Cocktail`, `Heapsort`, `Insertion`, `Merge`,
`Selection`, and `Reverse`. For ordinary application sorting, compare these
algorithms with the Go standard library's `slices.SortFunc` or `sort.Slice`.

### Order a dependency graph

The DFS implementation supports generic comparable keys. Add vertices before
adding edges; an edge `from -> to` means that `from` must come first.

```go
package main

import (
	"fmt"
	"log"

	"go.osspkg.com/algorithms/graph/dfs"
)

func main() {
	g := dfs.NewGraph[string]()
	for _, node := range []string{"clean", "compile", "test"} {
		if err := g.AddNode(node); err != nil {
			log.Fatal(err)
		}
	}
	if err := g.AddEdge("clean", "compile"); err != nil {
		log.Fatal(err)
	}
	if err := g.AddEdge("compile", "test"); err != nil {
		log.Fatal(err)
	}

	order, err := g.TopologicalSort()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(order)
}
```

`graph/kahn` is an alternative for string-keyed graphs. It registers vertices
automatically with `Add` and supports `BreakPoint` when only one target and its
transitive dependencies should be ordered.

### Encode values

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/encoding/base62"
)

func main() {
	codec := base62.New("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	encoded := codec.Encode(123456)
	decoded := codec.Decode(encoded)

	fmt.Println(encoded, decoded)
}
```

`base62.New` requires an alphabet of exactly 62 unique bytes. `Encode(0)` and
`Decode("")` both produce the zero value. Invalid input to `Decode` also
produces `0`, so validate input or use a round-trip check when zero must be
distinguished from invalid data.

For one-time passwords, use `encoding/otp`:

```go
package main

import (
	"fmt"
	"log"

	"go.osspkg.com/algorithms/encoding/otp"
)

func main() {
	generator, err := otp.New(otp.OptHashSHA256(), otp.OptCode8Digits())
	if err != nil {
		log.Fatal(err)
	}

	secret, err := generator.NewSecret(20)
	if err != nil {
		log.Fatal(err)
	}
	code, err := generator.GenerateTOTP(secret, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)
}
```

The default OTP configuration is SHA-1, a 30-second TOTP period, and 6-digit
codes. Treat secrets and `otpauth` URLs as sensitive credentials.

`encoding/token` provides compact `T64` identifiers:

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/encoding/token"
)

func main() {
	id := token.NewRandom()
	wire := id.String()
	parsed, ok := token.Parse(wire)
	if !ok {
		panic("invalid token")
	}
	fmt.Println(parsed == id)
}
```

`T64` also exposes marshal methods for text, binary, and JSON-oriented
serialization, plus `database/sql` scanner and valuer integration. Its current
`MarshalJSON` method returns the text representation without JSON quotes; wrap
the value as a string when standard JSON encoding is required.

### Use data structures

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/structs/bitmap"
	"go.osspkg.com/algorithms/trees/btree"
)

func main() {
	bm := bitmap.New()
	bm.Set(42)
	fmt.Println(bm.Has(42)) // true

	tree := btree.New[int, string](2)
	tree.Insert(10, "ten")
	value, ok := tree.Find(10)
	if ok {
		fmt.Println(value) // ten
	}
}
```

`structs/bloom` is useful when a fast membership check can tolerate false
positives. It never guarantees that a reported item is present, so keep a
source of truth behind the filter:

```go
package main

import (
	"log"

	"go.osspkg.com/algorithms/structs/bloom"
)

func main() {
	filter, err := bloom.New(bloom.Quantity(100_000, 0.01))
	if err != nil {
		log.Fatal(err)
	}
	filter.Add("user:42")
	possiblyPresent := filter.Contain("user:42")
	log.Println(possiblyPresent)
}
```

`structs/bitmap` and `trees/btree` use synchronization by default. Disable
bitmap locking only when access is strictly single-threaded or synchronized by
the caller.

## Testing

Run the complete test suite with:

```bash
go test ./...
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please open an issue for larger
changes and include tests, documentation, and an example when appropriate.

## License

Algorithms is released under the [BSD 3-Clause License](LICENSE).
