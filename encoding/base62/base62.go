/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package base62

import (
	"math"
)

const size = 62

type Base62 struct {
	enc []byte
	dec [256]byte
}

func New(alphabet string) *Base62 {
	if len(alphabet) != size {
		panic("encoding alphabet is not 62-bytes long")
	}
	v := &Base62{
		enc: []byte(alphabet),
	}
	for i := range v.dec {
		v.dec[i] = 0xff
	}
	for i, b := range v.enc {
		if v.dec[b] != 0xff {
			panic("encoding alphabet contains duplicate bytes")
		}
		v.dec[b] = byte(i)
	}
	return v
}

func (v *Base62) Encode(id uint64) string {
	var result [11]byte
	i := len(result)
	for id > 0 {
		i--
		result[i] = v.enc[id%size]
		id /= size
	}
	return string(result[i:])
}

func (v *Base62) Decode(data string) uint64 {
	var id uint64
	for i := 0; i < len(data); i++ {
		value := v.dec[data[i]]
		if value == 0xff || id > (math.MaxUint64-uint64(value))/size {
			return 0
		}
		id = id*size + uint64(value)
	}
	return id
}
