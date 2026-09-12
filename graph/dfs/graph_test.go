/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dfs

import (
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_DAG(t *testing.T) {
	dag := NewGraph[string]()

	// 1. Регистрируем ноды
	dag.AddNode("Clean")
	dag.AddNode("Compile")
	dag.AddNode("Test")
	dag.AddNode("Deploy")

	// 2. Строим связи (Clean -> Compile -> Test -> Deploy)
	_ = dag.AddEdge("Clean", "Compile")
	_ = dag.AddEdge("Compile", "Test")
	_ = dag.AddEdge("Test", "Deploy")

	// 3. Получаем линейный порядок выполнения
	order, err := dag.TopologicalSort()
	casecheck.NoError(t, err)
	casecheck.Equal(t, []string{"Clean", "Compile", "Test", "Deploy"}, order)

	// 4. Демонстрация детектора циклов
	_ = dag.AddEdge("Deploy", "Compile") // Создаем цикл: Compile -> Test -> Deploy -> Compile
	_, err = dag.TopologicalSort()
	casecheck.Error(t, err)
}

func BenchmarkTopologicalSort(b *testing.B) {
	const nodes = 1_000

	dag := NewGraph[int]()
	for i := 0; i < nodes; i++ {
		if err := dag.AddNode(i); err != nil {
			b.Fatal(err)
		}
		if i > 0 {
			if err := dag.AddEdge(i-1, i); err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := dag.TopologicalSort(); err != nil {
			b.Fatal(err)
		}
	}
}
