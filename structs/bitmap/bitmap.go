/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package bitmap

import (
	"fmt"
	"sync"
)

const (
	blockSize = 8
	MaxIndex  = uint64(1 << 34)
)

type Bitmap struct {
	bits []byte

	blocks  uint64
	max     uint64
	lockoff bool

	mux sync.RWMutex
}

type Option func(*Bitmap)

func OptDisableLock() Option {
	return func(o *Bitmap) {
		o.lockoff = true
	}
}

func OptMaxIndex(index uint64) Option {
	return func(o *Bitmap) {
		o.max = index
	}
}

func New(opts ...Option) *Bitmap {
	bm := &Bitmap{max: 1}

	for _, opt := range opts {
		opt(bm)
	}

	bm.resize(bm.max)

	return bm
}

func (b *Bitmap) resize(index uint64) {
	size := index/blockSize + 1

	b.bits = append(b.bits, make([]byte, size-b.blocks)...)
	b.blocks = uint64(len(b.bits))
	b.max = b.blocks*blockSize - 1
}

func (b *Bitmap) Set(index uint64) {
	if index > MaxIndex {
		return
	}

	if b.lockoff {
		if index > b.max {
			b.resize(index)
		}
		b.bits[index/blockSize] |= 1 << (index % blockSize)
		return
	}

	b.mux.Lock()
	if index > b.max {
		b.resize(index)
	}
	b.bits[index/blockSize] |= 1 << (index % blockSize)
	b.mux.Unlock()
}

func (b *Bitmap) Del(index uint64) {
	if index > b.max || index > MaxIndex {
		return
	}

	if b.lockoff {
		b.bits[index/blockSize] &^= 1 << (index % blockSize)
		return
	}

	b.mux.Lock()
	b.bits[index/blockSize] &^= 1 << (index % blockSize)
	b.mux.Unlock()
}

func (b *Bitmap) Has(index uint64) bool {
	if index > b.max || index > MaxIndex {
		return false
	}

	if b.lockoff {
		return (b.bits[index/blockSize] & (1 << (index % blockSize))) > 0
	}

	b.mux.RLock()
	result := (b.bits[index/blockSize] & (1 << (index % blockSize))) > 0
	b.mux.RUnlock()
	return result
}

func (b *Bitmap) MarshalBinary() ([]byte, error) {
	if !b.lockoff {
		b.mux.RLock()
		defer b.mux.RUnlock()
	}

	out := make([]byte, b.blocks)
	copy(out, b.bits)

	return out, nil
}

func (b *Bitmap) UnmarshalBinary(in []byte) error {
	maxBlocks := MaxIndex/blockSize + 1
	if len(in) == 0 {
		return fmt.Errorf("bitmap cannot be empty")
	}
	if uint64(len(in)) > maxBlocks {
		return fmt.Errorf("bitmap is too large")
	}

	if !b.lockoff {
		b.mux.Lock()
		defer b.mux.Unlock()
	}

	b.bits = make([]byte, len(in))
	copy(b.bits, in)

	b.blocks = uint64(len(in))
	b.max = b.blocks*blockSize - 1

	return nil
}

func (b *Bitmap) CopyTo(dst *Bitmap) {
	if b == dst {
		return
	}

	if !b.lockoff {
		b.mux.Lock()
	}

	bits := make([]byte, len(b.bits))
	copy(bits, b.bits)
	blocks, max, lockoff := b.blocks, b.max, b.lockoff

	if !b.lockoff {
		b.mux.Unlock()
	}

	if !dst.lockoff {
		dst.mux.Lock()
		defer dst.mux.Unlock()
	}
	dst.bits = bits
	dst.blocks = blocks
	dst.max = max
	dst.lockoff = lockoff
}
