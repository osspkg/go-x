/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

import "slices"

type damerauLevenshtein[T comparable] struct {
}

func NewDamerauLevenshtein[T comparable]() Comparer[T] {
	return &damerauLevenshtein[T]{}
}

func (l *damerauLevenshtein[T]) Distance(a, b []T) int {
	if slices.Equal(a, b) {
		return 0
	}

	n, m := len(a), len(b)

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}
	if n < m {
		b, a = a, b
		n, m = m, n
	}

	size := m + 1

	var buf [192]int
	var prev2, prev1, curr []int

	if size <= 64 {
		prev2 = buf[0:size]
		prev1 = buf[64 : 64+size]
		curr = buf[128 : 128+size]
	} else {
		heapBuf := make([]int, size*3)
		prev2 = heapBuf[0:size]
		prev1 = heapBuf[size : size*2]
		curr = heapBuf[size*2 : size*3]
	}

	for j := 0; j <= m; j++ {
		prev1[j] = j
	}

	for i := 1; i <= n; i++ {
		curr[0] = i

		aItem1 := a[i-1]
		var aItem2 T
		if i > 1 {
			aItem2 = a[i-2]
		}

		for j := 1; j <= m; j++ {
			bItem1 := b[j-1]

			cost := 0
			if aItem1 != bItem1 {
				cost = 1
			}

			curr[j] = min(
				prev1[j]+1,      // удаление
				curr[j-1]+1,     // вставка
				prev1[j-1]+cost, // замена
			)

			if i > 1 && j > 1 && aItem1 == b[j-2] && aItem2 == bItem1 {
				curr[j] = min(curr[j], prev2[j-2]+1) // транспозиция
			}
		}

		prev2, prev1, curr = prev1, curr, prev2
	}

	return prev1[m]
}

func (l *damerauLevenshtein[T]) Similarity(a, b []T) float64 {
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}

	dist := l.Distance(a, b)

	return round(1.0 - float64(dist)/float64(maxLen))
}
