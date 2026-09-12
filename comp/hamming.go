/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

type hamming[T comparable] struct {
}

func NewHamming[T comparable]() Comparer[T] {
	return &hamming[T]{}
}

func (h *hamming[T]) Distance(a, b []T) int {
	n, m := len(a), len(b)
	if n != m {
		return -1
	}

	dist := 0
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			dist++
		}
	}

	return dist
}

func (h *hamming[T]) Similarity(a, b []T) float64 {
	n, m := len(a), len(b)
	if n != m {
		return 0.0
	}

	if n == 0 {
		return 1.0
	}

	dist := 0
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			dist++
		}
	}

	return round(1.0 - float64(dist)/float64(n))
}
