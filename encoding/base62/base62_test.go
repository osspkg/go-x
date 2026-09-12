/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package base62

import (
	"math"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestEncode_EncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		id   uint64
		want string
	}{
		{name: "Case1", id: 0, want: ""},
		{name: "Case1", id: 1, want: "1"},
		{name: "Case1", id: 2, want: "2"},
		{name: "Case1", id: 3, want: "3"},
		{name: "Case1", id: 4, want: "4"},
		{name: "Case1", id: 5, want: "5"},
		{name: "Case1", id: 6, want: "6"},
		{name: "Case1", id: 7, want: "7"},
		{name: "Case1", id: 8, want: "8"},
		{name: "Case1", id: 9, want: "9"},
		{name: "Case2", id: 10, want: "Q"},
		{name: "Case3", id: 100, want: "1z"},
		{name: "Case4", id: 1000, want: "E8"},
		{name: "Case5", id: 10000, want: "2aC"},
		{name: "Case6", id: 100000, want: "H0m"},
		{name: "Case7", id: 1000000000, want: "15xjeE"},
		{name: "Case8", id: 999999, want: "4Z91"},
		{name: "Case20", id: math.MaxUint64, want: "VleDq16QDLX"},
	}

	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := v.Encode(tt.id)
			casecheck.Equal(t, tt.want, h)
			id := v.Decode(h)
			casecheck.Equal(t, tt.id, id)
		})
	}
}

func TestDecode_InvalidInput(t *testing.T) {
	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")
	if got := v.Decode("!"); got != 0 {
		t.Fatalf("invalid input decoded to %d", got)
	}
	if got := v.Decode("VVVVVVVVVVVVVVVVVVVV"); got != 0 {
		t.Fatalf("overflowing input decoded to %d", got)
	}
}

func TestCheckAll(t *testing.T) {
	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")

	for i := 0; i <= 1000; i++ {
		h := v.Encode(uint64(i))
		id := v.Decode(h)
		if uint64(i) != id {
			t.Errorf("case %d: want %d got %d", i, i, id)
		}
	}
}

func TestEncode_Encode(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want uint64
	}{
		{name: "Case1", str: "a", want: 37},
		{name: "Case2", str: "aa", want: 2331},
		{name: "Case3", str: "aaaaaaaa", want: 132435801748215},
	}

	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := v.Decode(tt.str)
			casecheck.Equal(t, tt.want, h)
		})
	}
}

func Benchmark_base62(b *testing.B) {
	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Decode(v.Encode(math.MaxUint64))
		}
	})
}

func Benchmark_base62_encode(b *testing.B) {
	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Encode(math.MaxUint64)
		}
	})
}

func Benchmark_base62_decode(b *testing.B) {
	v := New("0123456789QAZWSXEDCRFVTGBYHNUJMIKOLPqazwsxedcrfvtgbyhnujmikolp")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Decode("XNWjtpSoji4")
		}
	})
}
