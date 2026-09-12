/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"crypto/md5"
	crand "crypto/rand"
	"encoding/binary"
	"time"
)

type T64 [8]byte

var Nil T64

func (t T64) String() string {
	tableMu.RLock()
	defer tableMu.RUnlock()

	var dst [18]byte
	dlt := len(table)
	j := 0

	for _, v := range t {
		switch j {
		case 6, 11:
			dst[j] = '-'
			j++
			fallthrough
		default:
			dst[j] = table[int(v)/dlt]
			dst[j+1] = table[int(v)%dlt]
			j += 2
		}
	}

	return string(dst[:])
}

func (t T64) Uint64() uint64 {
	return binary.BigEndian.Uint64(t[:])
}

func ParseBytes(s []byte) (T64, bool) {
	tableMu.RLock()
	defer tableMu.RUnlock()

	n := len(s)
	if n != 18 || s[6] != '-' || s[11] != '-' {
		return Nil, false
	}

	dlt := len(table)
	var (
		t T64
		j = 0
	)

	for i := 0; i < n; i += 2 {
		if i == 6 || i == 11 {
			i--
			continue
		}
		a := reverseTable[s[i]]
		b := reverseTable[s[i+1]]
		if a == -1 || b == -1 {
			return Nil, false
		}
		t[j] = byte(a*dlt + b)
		j++
	}

	return t, true
}

func Parse(s string) (T64, bool) {
	tableMu.RLock()
	defer tableMu.RUnlock()

	if len(s) != 18 || s[6] != '-' || s[11] != '-' {
		return Nil, false
	}

	dlt := len(table)
	var t T64
	j := 0
	for i := 0; i < len(s); i += 2 {
		if i == 6 || i == 11 {
			i--
			continue
		}
		a := reverseTable[s[i]]
		b := reverseTable[s[i+1]]
		if a == -1 || b == -1 {
			return Nil, false
		}
		t[j] = byte(a*dlt + b)
		j++
	}

	return t, true
}

func NewByTime() (t T64) {
	now := time.Now().UnixNano()
	binary.BigEndian.PutUint64(t[:], uint64(now))
	return
}

func NewFormTime(tt time.Time) (t T64) {
	now := tt.UnixNano()
	binary.BigEndian.PutUint64(t[:], uint64(now))
	return
}

func NewByUint(v uint64) (t T64) {
	binary.BigEndian.PutUint64(t[:], v)
	return
}

func NewByBytes(b []byte) (t T64) {
	sum := md5.Sum(b) //nolint:gosec
	copy(t[:], sum[:])
	return
}

func NewByString(b string) T64 {
	return NewByBytes([]byte(b))
}

func NewRandom() (t T64) {
	for i := 0; i < 10; i++ {
		if _, err := crand.Read(t[:]); err != nil {
			continue
		}
		return
	}

	return
}
