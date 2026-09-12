/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package bitmap

import (
	"fmt"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_Bitmap_CopyTO(t *testing.T) {
	src := New()
	src.Set(1)
	src.Set(5)
	src.Set(60)

	dst := New()
	src.CopyTo(dst)

	casecheck.Equal(t, src.blocks, dst.blocks)
	casecheck.Equal(t, src.bits, dst.bits)
	casecheck.Equal(t, src.lockoff, dst.lockoff)
	casecheck.Equal(t, src.max, dst.max)
}

func TestUnit_Bitmap_Resize(t *testing.T) {
	t.Skip("Only for debug")

	src := New()

	for i := 0; i <= 20; i++ {
		src.Set(uint64(i))
		b, _ := src.MarshalBinary()
		fmt.Printf("(%d) %b `%s`\n", i, b, string(b))
	}
}

func TestUnit_Bitmap_Marshaling(t *testing.T) {
	bm := New()

	for i := 0; i <= 65; i++ {
		bm.Set(uint64(i))

		casecheck.True(t, bm.Has(uint64(i)), "(1) for index: %d", i)
		casecheck.False(t, bm.Has(uint64(i+1)), "(2) for index: %d", i+1)
	}

	backup, _ := bm.MarshalBinary()

	bm.UnmarshalBinary(make([]byte, len(backup)))

	for i := 0; i <= 65; i++ {
		casecheck.False(t, bm.Has(uint64(i)), "(3) for index: %d", i)
		casecheck.False(t, bm.Has(uint64(i+1)), "(4) for index: %d", i+1)
	}

	bm.UnmarshalBinary(backup)

	for i := 65; i >= 0; i-- {
		casecheck.True(t, bm.Has(uint64(i)), "(1) for index: %d", i)
		casecheck.False(t, bm.Has(uint64(i+1)), "(2) for index: %d", i+1)

		bm.Del(uint64(i))
		casecheck.False(t, bm.Has(uint64(i)), "(3) for index: %d", i+1)
	}

}

func TestUnit_Bitmap_UnmarshalBounds(t *testing.T) {
	bm := New()
	if err := bm.UnmarshalBinary([]byte{0}); err != nil {
		t.Fatal(err)
	}

	bm.Set(7)
	if !bm.Has(7) {
		t.Fatal("bit 7 should be set")
	}
	if bm.Has(8) {
		t.Fatal("bit 8 must be outside a one-byte bitmap")
	}
	bm.Set(8)
	if !bm.Has(8) {
		t.Fatal("bit 8 should resize the bitmap")
	}
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/algorithms/structs/bitmap
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_Bitmap
Benchmark_Bitmap-4   	 9515085	       182.2 ns/op	     225 B/op	       0 allocs/op
*/
func Benchmark_Bitmap(b *testing.B) {
	index := uint64(1 << 34)

	bm := New(OptMaxIndex(1024))

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bm.Set(index)
			bm.Has(index)
			bm.Del(index)
		}
	})
}
