/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comp

import "math"

type Comparer[T comparable] interface {
	Similarity(a, b []T) float64
}

func round(x float64) float64 {
	return math.Round(x*1_00_00) / 100
}
