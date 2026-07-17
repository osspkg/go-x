/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

type lcs[T comparable] struct {
}

func NewLCS[T comparable]() Comparer[T] {
	return &lcs[T]{}
}

func (l *lcs[T]) Distance(a, b []T) int {
	x, y := a, b
	n, m := len(x), len(y)

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}
	if n < m {
		y, x = x, y
		n, m = m, n
	}

	size := m + 1

	var buf [128]int
	var prev, curr []int

	if size <= 64 {
		prev = buf[0:size]
		curr = buf[64 : 64+size]
	} else {
		heapBuf := make([]int, size*2)
		prev = heapBuf[0:size]
		curr = heapBuf[size : size*2]
	}

	for i := 1; i <= n; i++ {
		aItem := x[i-1]

		for j := 1; j <= m; j++ {
			if aItem == y[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = max(prev[j], curr[j-1])
			}
		}
		prev, curr = curr, prev
	}

	return m + n - 2*prev[m]
}

func (l *lcs[T]) Similarity(a, b []T) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	return round((float64(len(a)+len(b)-l.Distance(a, b)) / 2) / float64(max(len(a), len(b))))
}
