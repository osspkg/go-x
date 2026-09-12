/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// see: https://en.wikipedia.org/wiki/Insertion_sort

package sorts

func Insertion[T any](list []T, less func(i, j int) bool) {
	if len(list) < 2 {
		return
	}
	for el := 1; el < len(list); el++ {
		for i := el; i > 0; i-- {
			if !less(i, i-1) {
				break
			}
			list[i-1], list[i] = list[i], list[i-1]
		}
	}
}
