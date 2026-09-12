/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package btree

import (
	"cmp"
	"sync"
)

type node[K cmp.Ordered, V any] struct {
	isLeaf   bool
	numKeys  int
	keys     []K
	values   []V
	children []*node[K, V]
}

func newNode[K cmp.Ordered, V any](size int, isLeaf bool) *node[K, V] {
	return &node[K, V]{
		isLeaf:   isLeaf,
		keys:     make([]K, size),
		values:   make([]V, size),
		children: make([]*node[K, V], size+1),
	}
}

type BTree[K cmp.Ordered, V any] struct {
	degree  int
	maxKeys int
	minKeys int
	root    *node[K, V]
	mu      sync.RWMutex
}

func New[K cmp.Ordered, V any](degree int) *BTree[K, V] {
	degree = max(degree, 2)

	b := &BTree[K, V]{
		degree:  degree,
		maxKeys: 2*degree - 1,
		minKeys: degree - 1,
	}

	b.root = newNode[K, V](b.maxKeys, true)

	return b
}

func (t *BTree[K, V]) Find(key K) (V, bool) {
	t.mu.RLock()

	curr := t.root
	for curr != nil {
		i := 0
		for i < curr.numKeys && key > curr.keys[i] {
			i++
		}

		if i < curr.numKeys && key == curr.keys[i] {
			value := curr.values[i]
			t.mu.RUnlock()
			return value, true
		}

		if curr.isLeaf {
			break
		}
		curr = curr.children[i]
	}

	var zero V
	t.mu.RUnlock()
	return zero, false
}

func (t *BTree[K, V]) Insert(key K, val V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if v := t.search(t.root, key); v != nil {
		*v = val
		return
	}

	root := t.root
	if root.numKeys == t.maxKeys {
		newRoot := newNode[K, V](t.maxKeys, false)
		t.root = newRoot
		newRoot.children[0] = root
		t.splitChild(newRoot, 0, root)
		t.insertNonFull(newRoot, key, val)
	} else {
		t.insertNonFull(root, key, val)
	}
}

func (t *BTree[K, V]) search(x *node[K, V], key K) *V {
	for x != nil {
		i := 0
		for i < x.numKeys && key > x.keys[i] {
			i++
		}
		if i < x.numKeys && key == x.keys[i] {
			return &x.values[i]
		}
		if x.isLeaf {
			break
		}
		x = x.children[i]
	}
	return nil
}

func (t *BTree[K, V]) splitChild(x *node[K, V], i int, y *node[K, V]) {
	z := newNode[K, V](t.maxKeys, y.isLeaf)
	z.numKeys = t.minKeys

	for j := 0; j < t.minKeys; j++ {
		z.keys[j] = y.keys[j+t.degree]
		z.values[j] = y.values[j+t.degree]
	}

	if !y.isLeaf {
		for j := 0; j < t.degree; j++ {
			z.children[j] = y.children[j+t.degree]
		}
	}

	y.numKeys = t.minKeys

	for j := x.numKeys; j >= i+1; j-- {
		x.children[j+1] = x.children[j]
	}
	x.children[i+1] = z

	for j := x.numKeys - 1; j >= i; j-- {
		x.keys[j+1] = x.keys[j]
		x.values[j+1] = x.values[j]
	}

	x.keys[i] = y.keys[t.minKeys]
	x.values[i] = y.values[t.minKeys]
	x.numKeys++
}

func (t *BTree[K, V]) insertNonFull(x *node[K, V], key K, val V) {
	i := x.numKeys - 1

	if x.isLeaf {
		for i >= 0 && key < x.keys[i] {
			x.keys[i+1] = x.keys[i]
			x.values[i+1] = x.values[i]
			i--
		}
		x.keys[i+1] = key
		x.values[i+1] = val
		x.numKeys++
	} else {
		for i >= 0 && key < x.keys[i] {
			i--
		}
		i++

		if x.children[i].numKeys == t.maxKeys {
			t.splitChild(x, i, x.children[i])
			if key > x.keys[i] {
				i++
			}
		}
		t.insertNonFull(x.children[i], key, val)
	}
}

func (t *BTree[K, V]) Delete(key K) {
	t.mu.Lock()
	defer t.mu.Unlock()

	root := t.root
	if root.numKeys == 0 {
		return
	}

	t.deleteFromNode(root, key)

	if root.numKeys == 0 && !root.isLeaf {
		t.root = root.children[0]
	}
}

func (t *BTree[K, V]) deleteFromNode(x *node[K, V], key K) {
	idx := 0
	for idx < x.numKeys && key > x.keys[idx] {
		idx++
	}

	if idx < x.numKeys && key == x.keys[idx] {
		if x.isLeaf {
			for i := idx; i < x.numKeys-1; i++ {
				x.keys[i] = x.keys[i+1]
				x.values[i] = x.values[i+1]
			}
			x.numKeys--
			var zeroK K
			var zeroV V
			x.keys[x.numKeys], x.values[x.numKeys] = zeroK, zeroV
		} else {
			t.deleteFromInternalNode(x, idx)
		}
		return
	}

	if x.isLeaf {
		return
	}

	isLastChild := idx == x.numKeys

	if x.children[idx].numKeys == t.minKeys {
		t.fill(x, idx)
	}

	if isLastChild && idx > x.numKeys {
		t.deleteFromNode(x.children[idx-1], key)
	} else {
		t.deleteFromNode(x.children[idx], key)
	}
}

func (t *BTree[K, V]) deleteFromInternalNode(x *node[K, V], idx int) {
	key := x.keys[idx]
	left := x.children[idx]
	right := x.children[idx+1]

	if left.numKeys >= t.degree {
		predKey, predVal := t.getPred(left)
		x.keys[idx] = predKey
		x.values[idx] = predVal
		t.deleteFromNode(left, predKey)
	} else if right.numKeys >= t.degree {
		succKey, succVal := t.getSucc(right)
		x.keys[idx] = succKey
		x.values[idx] = succVal
		t.deleteFromNode(right, succKey)
	} else {
		t.merge(x, idx)
		t.deleteFromNode(left, key)
	}
}

func (t *BTree[K, V]) getPred(x *node[K, V]) (K, V) {
	curr := x
	for !curr.isLeaf {
		curr = curr.children[curr.numKeys]
	}
	return curr.keys[curr.numKeys-1], curr.values[curr.numKeys-1]
}

func (t *BTree[K, V]) getSucc(x *node[K, V]) (K, V) {
	curr := x
	for !curr.isLeaf {
		curr = curr.children[0]
	}
	return curr.keys[0], curr.values[0]
}

func (t *BTree[K, V]) fill(x *node[K, V], idx int) {
	if idx != 0 && x.children[idx-1].numKeys >= t.degree {
		t.borrowFromPrev(x, idx)
	} else if idx != x.numKeys && x.children[idx+1].numKeys >= t.degree {
		t.borrowFromNext(x, idx)
	} else {
		if idx != x.numKeys {
			t.merge(x, idx)
		} else {
			t.merge(x, idx-1)
		}
	}
}

func (t *BTree[K, V]) borrowFromPrev(x *node[K, V], idx int) {
	child := x.children[idx]
	sibling := x.children[idx-1]

	for i := child.numKeys - 1; i >= 0; i-- {
		child.keys[i+1] = child.keys[i]
		child.values[i+1] = child.values[i]
	}

	if !child.isLeaf {
		for i := child.numKeys; i >= 0; i-- {
			child.children[i+1] = child.children[i]
		}
		child.children[0] = sibling.children[sibling.numKeys]
	}

	child.keys[0] = x.keys[idx-1]
	child.values[0] = x.values[idx-1]
	x.keys[idx-1] = sibling.keys[sibling.numKeys-1]
	x.values[idx-1] = sibling.values[sibling.numKeys-1]

	child.numKeys++
	sibling.numKeys--
	var zeroK K
	var zeroV V
	sibling.keys[sibling.numKeys], sibling.values[sibling.numKeys] = zeroK, zeroV
	if !sibling.isLeaf {
		sibling.children[sibling.numKeys+1] = nil
	}
}

func (t *BTree[K, V]) borrowFromNext(x *node[K, V], idx int) {
	child := x.children[idx]
	sibling := x.children[idx+1]

	child.keys[child.numKeys] = x.keys[idx]
	child.values[child.numKeys] = x.values[idx]

	if !child.isLeaf {
		child.children[child.numKeys+1] = sibling.children[0]
	}

	x.keys[idx] = sibling.keys[0]
	x.values[idx] = sibling.values[0]

	for i := 1; i < sibling.numKeys; i++ {
		sibling.keys[i-1] = sibling.keys[i]
		sibling.values[i-1] = sibling.values[i]
	}

	if !sibling.isLeaf {
		for i := 1; i <= sibling.numKeys; i++ {
			sibling.children[i-1] = sibling.children[i]
		}
	}

	child.numKeys++
	sibling.numKeys--
	var zeroK K
	var zeroV V
	sibling.keys[sibling.numKeys], sibling.values[sibling.numKeys] = zeroK, zeroV
	if !sibling.isLeaf {
		sibling.children[sibling.numKeys+1] = nil
	}
}

func (t *BTree[K, V]) merge(x *node[K, V], idx int) {
	child := x.children[idx]
	sibling := x.children[idx+1]

	child.keys[t.minKeys] = x.keys[idx]
	child.values[t.minKeys] = x.values[idx]

	for i := 0; i < sibling.numKeys; i++ {
		child.keys[i+t.degree] = sibling.keys[i]
		child.values[i+t.degree] = sibling.values[i]
	}

	if !child.isLeaf {
		for i := 0; i <= sibling.numKeys; i++ {
			child.children[i+t.degree] = sibling.children[i]
		}
	}

	for i := idx + 1; i < x.numKeys; i++ {
		x.keys[i-1] = x.keys[i]
		x.values[i-1] = x.values[i]
	}
	for i := idx + 2; i <= x.numKeys; i++ {
		x.children[i-1] = x.children[i]
	}

	child.numKeys += sibling.numKeys + 1
	x.numKeys--
	var zeroK K
	var zeroV V
	x.keys[x.numKeys], x.values[x.numKeys] = zeroK, zeroV
	x.children[x.numKeys+1] = nil
}
