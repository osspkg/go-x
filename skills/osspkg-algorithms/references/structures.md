# Data structures

## Bitmap: `structs/bitmap`

```go
package main

import (
	"log"

	"go.osspkg.com/algorithms/structs/bitmap"
)

func main() {
	bm := bitmap.New(bitmap.OptMaxIndex(1024))
	bm.Set(42)
	if !bm.Has(42) {
		log.Fatal("bit was not set")
	}
	bm.Del(42)

	data, err := bm.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}
	copyBM := bitmap.New()
	if err := copyBM.UnmarshalBinary(data); err != nil {
		log.Fatal(err)
	}
}
```

`Set`, `Del`, and `Has` use `uint64` indices; the bitmap grows automatically to
the requested index, but values above the exported `bitmap.MaxIndex` are
ignored or treated as absent. `MarshalBinary` returns a copy of the raw bytes.
`UnmarshalBinary` rejects empty input.

Operations are protected by an `RWMutex` by default. Use `OptDisableLock()` only
when access is strictly single-threaded or synchronization is fully handled
externally; it is not a safe default optimization.

## Bloom filter: `structs/bloom`

```go
package main

import (
	"bytes"
	"log"

	"go.osspkg.com/algorithms/structs/bloom"
)

func main() {
	filter, err := bloom.New(bloom.Quantity(100_000, 0.01))
	if err != nil {
		log.Fatal(err)
	}
	filter.Add("user:42")
	if !filter.Contain("user:42") {
		log.Fatal("inserted value is absent")
	}

	var buf bytes.Buffer
	if err := filter.Dump(&buf); err != nil {
		log.Fatal(err)
	}

	restored, err := bloom.New(bloom.Quantity(100_000, 0.01))
	if err != nil {
		log.Fatal(err)
	}
	if err := restored.Restore(bytes.NewReader(buf.Bytes())); err != nil {
		log.Fatal(err)
	}
}
```

`Quantity(expectedItems, falsePositiveRate)` sets the expected number of items
and the target false-positive probability; the size must be greater than zero,
and the rate must be strictly between `0` and `1`. xxhash is used by default.
With `HashFunc`, provide a `func() hash.Hash` factory such as `sha256.New`.

The filter answers "definitely not" or "possibly present": false positives are
normal, and `Contain` does not replace the source of truth. Use the same key
representation with `Add` and `Contain`; the library converts strings, bytes,
integers, and values implementing supported interfaces into bytes.

`Dump` stores the salts and bitmap, but does not store the hash function or the
original `Quantity` for the consuming code. To restore, use the same quantity
configuration and hash factory as when the filter was written, then check the
error from `Restore`. `CopyTo` copies the filter state independently of the
source.

## B-tree: `trees/btree`

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/trees/btree"
)

func main() {
	tree := btree.New[int, string](2)
	tree.Insert(10, "ten")
	tree.Insert(20, "twenty")
	tree.Insert(10, "TEN") // duplicate key updates the value

	if value, ok := tree.Find(10); ok {
		fmt.Println(value) // TEN
	}
	tree.Delete(20)
}
```

Key `K` must satisfy `cmp.Ordered`, while value `V` can be any type. `degree < 2`
is automatically replaced with `2`. `Insert` updates the value for a duplicate
key; `Find` returns `(zeroValue, false)` for a missing key, so always check the
`bool`. `Delete` does nothing for a missing key. There is no public traversal or
iterator.

`BTree` protects `Find`, `Insert`, and `Delete` with an internal `RWMutex`, but
do not copy the value after creation: pass around the pointer returned by `New`.

Verified examples in the checkout: `structs/bitmap/bitmap_test.go`,
`structs/bloom/bloom_test.go`, and `trees/btree/btree_test.go`.
