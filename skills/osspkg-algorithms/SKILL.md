---
name: osspkg-algorithms
description: Use the go.osspkg.com/algorithms Go module for sequence comparison, sorting, graph ordering, encodings, probabilistic filters, bitmaps, and B-trees; read the package-specific references for exact APIs and behavior.
metadata:
  short-description: Use OSSPkg Go algorithms
---

# OSSPkg Algorithms

Use this skill when writing or reviewing Go code that imports
`go.osspkg.com/algorithms`. The library has no useful root package: import the
specific subpackage you need.

## Routing

Load only the relevant reference:

- [comparisons.md](references/comparisons.md) — `comp`: string similarity and arbitrary comparable sequences.
- [sorting-and-graphs.md](references/sorting-and-graphs.md) — `sorts`, `graph/dfs`, `graph/kahn`.
- [encoding.md](references/encoding.md) — `encoding/base62`, `encoding/otp`, `encoding/token`.
- [structures.md](references/structures.md) — `structs/bitmap`, `structs/bloom`, `trees/btree`.

## General rules

- Use imports such as `go.osspkg.com/algorithms/<subpackage>` and explicit type
  arguments when a generic constructor cannot infer them from its arguments:
  `comp.NewLevenshtein[rune]()`, `dfs.NewGraph[string]()` and
  `btree.New[int, string](2)`.
- Check the signature of the public constructor itself. For example, `comp`
  returns a `Comparer` interface whose only public method is `Similarity`; the
  internal `Distance` method is not part of the external API.
- Handle every returned error. Also account for APIs that report failure with a
  `bool` or return a zero value instead of an `error`.
- Sorting functions mutate the provided slice in place, and the comparison
  callback receives indices into the current slice, not the values themselves.
- Before implementing new behavior, check the library's current source and
  tests: these references document API usage, not hidden implementation details.

For a dependency in another Go module, use `go get go.osspkg.com/algorithms`.
In this checkout, verify changes in the consuming package or test and run
`go test ./...`.
