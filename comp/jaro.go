/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

import "slices"

type jaro[T comparable] struct {
}

func NewJaroWinkler[T comparable]() Comparer[T] {
	return &jaro[T]{}
}

func (l *jaro[T]) Similarity(a, b []T) float64 {
	if slices.Equal(a, b) {
		return 100.0
	}

	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	r1, r2 := a, b
	l1, l2 := len(r1), len(r2)

	matchDistance := l1
	if l2 > matchDistance {
		matchDistance = l2
	}
	matchDistance = (matchDistance / 2) - 1
	if matchDistance < 0 {
		matchDistance = 0
	}

	var m1Buf, m2Buf [64]bool
	var m1, m2 []bool
	if l1 <= 64 {
		m1 = m1Buf[:l1]
	} else {
		m1 = make([]bool, l1)
	}
	if l2 <= 64 {
		m2 = m2Buf[:l2]
	} else {
		m2 = make([]bool, l2)
	}

	matches := 0
	for i := 0; i < l1; i++ {
		start := i - matchDistance
		if start < 0 {
			start = 0
		}
		end := i + matchDistance + 1
		if end > l2 {
			end = l2
		}

		for j := start; j < end; j++ {
			if m2[j] {
				continue
			}
			if r1[i] == r2[j] {
				m1[i] = true
				m2[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0.0
	}

	t := 0
	k := 0
	for i := 0; i < l1; i++ {
		if !m1[i] {
			continue
		}
		for !m2[k] {
			k++
		}
		if r1[i] != r2[k] {
			t++
		}
		k++
	}
	t /= 2

	m := float64(matches)
	return round((m/float64(l1) + m/float64(l2) + (m-float64(t))/m) / 3.0)
}
