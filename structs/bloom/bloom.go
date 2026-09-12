/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// see: https://en.wikipedia.org/wiki/Bloom_filter

// Фильтр Блума - это пространственно-эффективная вероятностная структура данных,
// созданная для проверки наличия элемента в множестве. Он спроектирован невероятно
// быстрым при минимальном использовании памяти ценой потенциальных ложных срабатываний.
// Существует возможность получить ложноположительное срабатывание
// (элемента в множестве нет, но структура данных сообщает, что он есть),
// но не ложноотрицательное. Другими словами, очередь возвращает или "возможно в наборе", или "определённо не в наборе".
// Фильтр Блума может использовать любой объём памяти, однако чем он больше, тем меньше вероятность ложного срабатывания.

package bloom

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"math"
	"strconv"
	"sync"

	"github.com/cespare/xxhash/v2"

	"go.osspkg.com/algorithms/structs/bitmap"
)

const saltSize = 8

const (
	maxRestoreBytes = 64 << 20
	maxRestoreSalts = 1 << 20
)

type Bloom struct {
	bits  *bitmap.Bitmap
	size  uint64
	salts [][saltSize]byte

	optSize uint64
	optRate float64

	pool *sync.Pool
	mux  sync.RWMutex
}

type Option func(b *Bloom)

func HashFunc(h func() hash.Hash) Option {
	return func(b *Bloom) {
		b.pool = &sync.Pool{New: func() any { return h() }}
	}
}

func Quantity(size uint64, rate float64) Option {
	return func(b *Bloom) {
		b.optSize = size
		b.optRate = rate
	}
}

func New(opts ...Option) (*Bloom, error) {
	b := &Bloom{
		optSize: 10_000_000,
		optRate: 0.1,
		pool:    &sync.Pool{New: func() any { return xxhash.New() }},
	}

	for _, opt := range opts {
		opt(b)
	}

	if b.optSize == 0 {
		return nil, fmt.Errorf("bitset size cannot be 0")
	}
	if b.optRate <= 0.0 || b.optRate >= 1.0 {
		return nil, fmt.Errorf("false positive rate must be between 0.0 and 1.0")
	}

	m, k := calcOptimalParams(b.optSize, b.optRate)

	b.size = m
	b.bits = bitmap.New(bitmap.OptMaxIndex(m), bitmap.OptDisableLock())
	b.salts = make([][saltSize]byte, k)

	for i := 0; i < int(k); i++ {
		if _, err := rand.Read(b.salts[i][:]); err != nil {
			return nil, fmt.Errorf("generate hash salt: %w", err)
		}

		b.salts[i] = [saltSize]byte(bytes.ReplaceAll(b.salts[i][:], []byte("\n"), []byte("~")))
	}

	return b, nil
}

func (b *Bloom) CopyTo(dst *Bloom) {
	if b == dst {
		return
	}

	b.mux.RLock()
	bits := bitmap.New(bitmap.OptDisableLock())
	b.bits.CopyTo(bits)
	size, optSize, optRate := b.size, b.optSize, b.optRate
	salts := append([][saltSize]byte(nil), b.salts...)
	b.mux.RUnlock()

	dst.mux.Lock()
	defer dst.mux.Unlock()
	dst.bits = bits
	dst.size = size
	dst.salts = salts
	dst.optSize = optSize
	dst.optRate = optRate
}

func (b *Bloom) Dump(w io.Writer) error {
	b.mux.RLock()
	defer b.mux.RUnlock()

	if _, err := w.Write([]byte("OSSPkg:bloom\n")); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if _, err := fmt.Fprintf(w, "%d\n", len(b.salts)); err != nil {
		return fmt.Errorf("write salt count: %w", err)
	}

	for _, salt := range b.salts {
		if _, err := w.Write(salt[:]); err != nil {
			return fmt.Errorf("write salt: %w", err)
		}

		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("write salt: %w", err)
		}
	}

	bb, err := b.bits.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal bitset: %w", err)
	}

	if _, err = w.Write(bb); err != nil {
		return fmt.Errorf("write bitmap: %w", err)
	}

	return nil
}

func (b *Bloom) Restore(r io.Reader) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	reader := bufio.NewReader(io.LimitReader(r, maxRestoreBytes+1))

	head, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	if !bytes.Equal(head[:len(head)-1], []byte("OSSPkg:bloom")) {
		return fmt.Errorf("invalid header")
	}

	countSalt, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read countSalt: %w", err)
	}

	count, err := strconv.Atoi(string(countSalt[:len(countSalt)-1]))
	if err != nil {
		return fmt.Errorf("invalid countSalt: %w", err)
	}

	if count <= 0 || count > maxRestoreSalts {
		return fmt.Errorf("invalid countSalt")
	}

	salts := make([][saltSize]byte, count)

	for i := 0; i < count; i++ {
		salt, err0 := reader.ReadBytes('\n')
		if err0 != nil {
			return fmt.Errorf("read salt[%d]: %w", i, err0)
		}

		salt = salt[:len(salt)-1]
		if len(salt) != saltSize {
			return fmt.Errorf("invalid salt[%d], want 64 got %d", i, len(salt))
		}

		salts[i] = [saltSize]byte(salt)
	}

	bm, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read bitmap: %w", err)
	}
	if len(bm) == 0 || len(bm) > maxRestoreBytes {
		return fmt.Errorf("invalid bitmap size")
	}
	if uint64(len(bm)) > bitmap.MaxIndex/8+1 {
		return fmt.Errorf("bitmap is too large")
	}

	bits := bitmap.New(bitmap.OptDisableLock())
	if err := bits.UnmarshalBinary(bm); err != nil {
		return fmt.Errorf("restore bitmap: %w", err)
	}
	b.bits = bits
	b.salts = salts
	return nil
}

func (b *Bloom) Add(arg any) {
	h, ok := b.pool.Get().(hash.Hash)
	if !ok {
		panic("failed get hash function from pool")
	}
	defer func() {
		b.pool.Put(h)
	}()

	val := anyToBytes(arg)

	b.mux.Lock()
	defer b.mux.Unlock()
	sum := make([]byte, 0, h.Size())

	for i := 0; i < len(b.salts); i++ {
		h.Reset()
		h.Write(val)
		h.Write(b.salts[i][:])
		sum = h.Sum(sum[:0])
		key := binary.BigEndian.Uint64(sum) % b.size

		b.bits.Set(key)
	}
}

func (b *Bloom) Contain(arg any) bool {
	h, ok := b.pool.Get().(hash.Hash)
	if !ok {
		panic("failed get hash function from pool")
	}
	defer func() {
		b.pool.Put(h)
	}()

	val := anyToBytes(arg)

	b.mux.RLock()
	defer b.mux.RUnlock()
	sum := make([]byte, 0, h.Size())

	for i := 0; i < len(b.salts); i++ {
		h.Reset()
		h.Write(val)
		h.Write(b.salts[i][:])
		sum = h.Sum(sum[:0])
		key := binary.BigEndian.Uint64(sum) % b.size

		if !b.bits.Has(key) {
			return false
		}
	}
	return true
}

func calcOptimalParams(n uint64, p float64) (uint64, uint64) {
	m := -(float64(n) * math.Log(p)) / math.Pow(math.Log(2.0), 2.0)
	if m < 1 {
		m = 1.0
	}
	k := (m / float64(n)) * math.Log(2.0)
	if k < 1 {
		k = 1.0
	}
	return uint64(math.Ceil(m)), uint64(math.Ceil(k))
}

type byter interface {
	Bytes() []byte
}

func anyToBytes(arg any) []byte {
	switch value := arg.(type) {
	case []byte:
		return value
	case byter:
		return value.Bytes()
	case string:
		return []byte(value)
	case fmt.Stringer:
		return []byte(value.String())
	case int64:
		return binary.AppendVarint(nil, value)
	case int32:
		return binary.AppendVarint(nil, int64(value))
	case int16:
		return binary.AppendVarint(nil, int64(value))
	case int8:
		return binary.AppendVarint(nil, int64(value))
	case int:
		return binary.AppendVarint(nil, int64(value))
	case uint64:
		return binary.AppendUvarint(nil, value)
	case uint32:
		return binary.AppendUvarint(nil, uint64(value))
	case uint16:
		return binary.AppendUvarint(nil, uint64(value))
	case uint8:
		return binary.AppendUvarint(nil, uint64(value))
	case uint:
		return binary.AppendUvarint(nil, uint64(value))
	case json.Marshaler:
		bb, _ := value.MarshalJSON()
		return bb
	case encoding.BinaryMarshaler:
		bb, _ := value.MarshalBinary()
		return bb
	case encoding.TextMarshaler:
		bb, _ := value.MarshalText()
		return bb
	case gob.GobEncoder:
		bb, _ := value.GobEncode()
		return bb
	default:
		return []byte(fmt.Sprintf("%+v", arg))
	}
}
