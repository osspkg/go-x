/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"errors"
	"sync"
)

var (
	table        []byte
	reverseTable [255]int
	tableMu      sync.RWMutex
)

func init() {
	if err := SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"); err != nil {
		panic(err)
	}
}

func SetTable(s string) error {
	if len(s)%2 != 0 {
		return errors.New("token: table must be a multiple of 2")
	}
	if len(s) < 16 {
		return errors.New("token: table must be 16 or more in length")
	}
	if len(s) > 255 {
		return errors.New("token: table should be no longer than 255")
	}

	uniq := make(map[byte]struct{}, len(s))

	t, r := []byte(s), [255]int{}
	for i := 0; i < len(r); i++ {
		r[i] = -1
	}
	for i, b := range t {
		if b == 0xff {
			return errors.New("token: table cannot contain byte 0xff")
		}
		if _, ok := uniq[b]; ok {
			return errors.New("token: duplicate chars in table")
		}
		uniq[b] = struct{}{}
		r[b] = i
	}
	tableMu.Lock()
	table, reverseTable = t, r
	tableMu.Unlock()
	return nil
}
