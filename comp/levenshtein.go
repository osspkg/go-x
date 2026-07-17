/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

type levenshtein[T comparable] struct {
}

func NewLevenshtein[T comparable]() Comparer[T] {
	return &levenshtein[T]{}
}

func (l *levenshtein[T]) Distance(a, b []T) int {
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

	var buf [64]int
	var col []int

	if size <= 64 {
		col = buf[:size]
	} else {
		col = make([]int, size)
	}

	for j := 0; j < size; j++ {
		col[j] = j
	}

	for i := 1; i <= n; i++ {
		prevDiag := col[0]
		col[0] = i

		aItem := a[i-1]

		for j := 1; j <= m; j++ {
			currDiag := col[j]

			cost := 0
			if aItem != b[j-1] {
				cost = 1
			}

			minVal := col[j] + 1 // удаление

			if ins := col[j-1] + 1; ins < minVal {
				minVal = ins // вставка
			}
			if sub := prevDiag + cost; sub < minVal {
				minVal = sub // замена
			}

			col[j] = minVal
			prevDiag = currDiag
		}
	}

	return col[m]
}

func (l *levenshtein[T]) Similarity(a, b []T) float64 {
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}

	dist := l.Distance(a, b)

	return round(1.0 - float64(dist)/float64(maxLen))
}
