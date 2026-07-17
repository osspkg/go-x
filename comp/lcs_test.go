/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

import (
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_NewLCS(t *testing.T) {
	{
		c := NewLCS[rune]()

		casecheck.Equal(t, 81.82, c.Similarity([]rune("Hello world"), []rune("Hello-World")))
		casecheck.Equal(t, 100.0, c.Similarity([]rune("Hello world"), []rune("Hello world")))
		casecheck.Equal(t, 45.45, c.Similarity([]rune("Hello world"), []rune("world Hello")))
		casecheck.Equal(t, 90.91, c.Similarity([]rune("Hello world"), []rune("Helle world")))
	}

	{
		c := NewLCS[int]()

		casecheck.Equal(t, 100.0, c.Similarity([]int{1, 2, 3}, []int{1, 2, 3}))
		casecheck.Equal(t, 66.67, c.Similarity([]int{1, 2, 3}, []int{1, 3, 2}))
		casecheck.Equal(t, 0.0, c.Similarity([]int{1, 2, 3}, []int{4, 5, 6}))
	}
}

func Benchmark_NewLCS(b *testing.B) {
	c := NewLCS[rune]()
	m1, m2 := []rune("Hello world"), []rune("Hello-World")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Similarity(m1, m2)
		}
	})
}
