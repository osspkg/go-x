/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// see: https://en.wikipedia.org/wiki/Merge_sort

package sorts

func Merge[T any](list []T, less func(i, j int) bool) {
	tmp := make([]T, 0, len(list))
	mergeSplitSortJoin[T](list, tmp, 0, len(list)-1, less)
}

func mergeSplitSortJoin[T any](list, tmp []T, from, to int, less func(i, j int) bool) {
	size := to - from + 1
	switch true {
	case size <= 1:
		return
	case size == 2:
		if less(to, from) {
			list[to], list[from] = list[from], list[to]
		}
		return
	}

	half := size / 2

	mergeSplitSortJoin[T](list, tmp, from+half, to, less)
	mergeSplitSortJoin[T](list, tmp, from, from+half-1, less)
	mergeJoin[T](list, tmp[:0], from, from+half-1, from+half, to, less)
}

func mergeJoin[T any](list, tmp []T, fromA, toA, fromB, toB int, less func(i, j int) bool) {
	nA, nB := fromA, fromB
	for {
		if nA <= toA && nB <= toB {
			if less(nA, nB) {
				tmp = append(tmp, list[nA])
				nA++
			} else {
				tmp = append(tmp, list[nB])
				nB++
			}
			continue
		}

		if nB > toB {
			for i := nA; i <= toA; i++ {
				tmp = append(tmp, list[i])
			}
		} else {
			for i := nB; i <= toB; i++ {
				tmp = append(tmp, list[i])
			}
		}
		break
	}
	for i, v := range tmp {
		list[fromA+i] = v
	}
}
