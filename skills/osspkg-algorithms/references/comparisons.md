# Sequence comparison: `comp`

Import:

```go
import "go.osspkg.com/algorithms/comp"
```

Public constructors:

| Constructor | Suitable use case |
| --- | --- |
| `NewHamming[T]()` | equal-length sequences; counts positional mismatches |
| `NewLevenshtein[T]()` | insertions, deletions, and substitutions |
| `NewDamerauLevenshtein[T]()` | the same, plus adjacent transpositions |
| `NewLCS[T]()` | similarity based on the longest common subsequence |
| `NewJaroWinkler[T]()` | short sequences and near matches |

`T` must satisfy `comparable`. The constructors return `comp.Comparer[T]`, so
external code calls `Similarity`:

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

`Similarity` returns a value on a `0..100` scale, rounded to two decimal places;
two equal empty sequences return `100`. For text, use `[]rune` when comparison
should operate on Unicode code points, and `[]byte` when the comparison must be
byte-oriented.

Selection notes:

- `Hamming` is intended for equal-length inputs; with different lengths its
  `Similarity` returns `0`.
- `Levenshtein` fits typo matching where insertions, deletions, and substitutions
  matter.
- `Damerau-Levenshtein` is useful for swapped adjacent characters.
- `LCS` better reflects the preserved order of a common subsequence.
- `Jaro-Winkler` is useful for short values; calibrate thresholds on domain data.

Do not call `Distance` on a constructor result from an external package:
implementations have internal distance methods, but their types are unexported
and `Comparer` does not declare that method.

Verified examples in the checkout: `comp/*_test.go`.
