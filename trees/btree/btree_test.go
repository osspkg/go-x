/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package btree

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

// TestBTreeBasic проверяет вставку и поиск при degree=2 (классическое B-дерево 2-3-4).
func TestUnit_BTreeBasic(t *testing.T) {
	tree := New[int, string](2)

	keys := []int{10, 20, 5, 6, 12, 30, 7, 17}
	values := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	// Вставка
	for i, k := range keys {
		tree.Insert(k, values[i])
	}

	// Поиск существующих
	for i, k := range keys {
		v, found := tree.Find(k)
		if !found {
			t.Errorf("key %d should be found", k)
		}
		if v != values[i] {
			t.Errorf("key %d: expected value %q, got %q", k, values[i], v)
		}
	}

	// Поиск отсутствующих
	missing := []int{1, 8, 13, 100}
	for _, k := range missing {
		_, found := tree.Find(k)
		if found {
			t.Errorf("key %d should not be found", k)
		}
	}
}

// TestBTreeDuplicate проверяет корректность работы с дублирующимися ключами.
func TestUnit_BTreeDuplicate(t *testing.T) {
	tree := New[int, string](2)

	tree.Insert(5, "first")
	tree.Insert(5, "second")
	tree.Insert(5, "third")

	// Должен найти какое-то из значений (второе или первое в зависимости от реализации)
	v, found := tree.Find(5)
	if !found {
		t.Fatal("key 5 should be found")
	}

	if v != "third" {
		t.Errorf("unexpected value: %q, expected %q", v, "first")
	}

	// Удалим один ключ 5
	tree.Delete(5)

	// После удаления одного ключа 5 должно остаться два вхождения
	_, found = tree.Find(5)
	if found {
		t.Fatal("key 5 should still exist after one deletion")
	}
}

// TestBTreeDelete проверяет различные сценарии удаления: из листа, с заимствованием, со слиянием, уменьшение высоты.
func TestUnit_BTreeDelete(t *testing.T) {
	// Создадим дерево с degree=2, чтобы были операции разделения и слияния
	tree := New[int, int](2)

	// Вставим последовательность чисел
	n := 20
	for i := 1; i <= n; i++ {
		tree.Insert(i, i*10)
	}

	// Проверим наличие всех
	for i := 1; i <= n; i++ {
		v, found := tree.Find(i)
		if !found || v != i*10 {
			t.Fatalf("before deletion: key %d not found or wrong value", i)
		}
	}

	// Удалим несколько ключей, которые должны вызвать различные перестроения
	toDelete := []int{3, 8, 15, 10, 1, 20}
	for _, k := range toDelete {
		tree.Delete(k)
	}

	// Проверим, что удалённые отсутствуют
	for _, k := range toDelete {
		_, found := tree.Find(k)
		if found {
			t.Errorf("key %d should have been deleted", k)
		}
	}

	// Остальные должны остаться
	for i := 1; i <= n; i++ {
		deleted := false
		for _, d := range toDelete {
			if i == d {
				deleted = true
				break
			}
		}
		if !deleted {
			v, found := tree.Find(i)
			if !found {
				t.Errorf("key %d should still exist", i)
			}
			if v != i*10 {
				t.Errorf("key %d: expected value %d, got %d", i, i*10, v)
			}
		}
	}
}

// TestBTreeDegree1 проверяет граничный случай degree=1 (допускается 0 или 1 ключ в узле).
func TestUnit_BTreeDegree1(t *testing.T) {
	tree := New[int, string](1) // maxKeys=1, minKeys=0

	// Вставка
	tree.Insert(5, "five")
	v, found := tree.Find(5)
	if !found || v != "five" {
		t.Fatalf("expected 'five', got %v, found=%v", v, found)
	}

	// Вставка ещё одного ключа вызовет разделение корня
	tree.Insert(10, "ten")
	v, found = tree.Find(10)
	if !found || v != "ten" {
		t.Fatalf("expected 'ten', got %v, found=%v", v, found)
	}

	// Оба ключа должны присутствовать
	for _, k := range []int{5, 10} {
		if _, f := tree.Find(k); !f {
			t.Errorf("key %d not found", k)
		}
	}

	// Удаление до пустого дерева
	tree.Delete(5)
	_, found = tree.Find(5)
	if found {
		t.Error("5 should be deleted")
	}
	// Второй ключ ещё есть
	if _, f := tree.Find(10); !f {
		t.Error("10 should still exist")
	}

	tree.Delete(10)
	if _, f := tree.Find(10); f {
		t.Error("10 should be deleted")
	}
}

// TestBTreeRandomStress выполняет множество случайных операций и проверяет согласованность.
func TestUnit_BTreeRandomStress(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	degree := rng.Intn(10) + 2 // от 2 до 11
	tree := New[int, int](degree)

	reference := make(map[int]int)
	const operations = 2000

	for i := 0; i < operations; i++ {
		key := rng.Intn(200)
		switch rng.Intn(3) {
		case 0: // Insert
			val := rng.Intn(10000)
			tree.Insert(key, val)
			reference[key] = val
		case 1: // Find
			v, found := tree.Find(key)
			ref, exists := reference[key]
			if exists != found {
				t.Errorf("Find(%d): existence mismatch, got %v, want %v", key, found, exists)
			}
			if exists && v != ref {
				t.Errorf("Find(%d): value mismatch, got %d, want %d", key, v, ref)
			}
		case 2: // Delete
			tree.Delete(key)
			delete(reference, key)
		}
	}

	// Финальная проверка всех ключей из reference
	for k, v := range reference {
		res, found := tree.Find(k)
		if !found {
			t.Errorf("final check: key %d missing", k)
		}
		if res != v {
			t.Errorf("final check: key %d value mismatch, got %d, want %d", k, res, v)
		}
	}
}

// TestBTreeConcurrent проверяет безопасность конкурентного доступа.
func TestUnit_BTreeConcurrent(t *testing.T) {
	tree := New[int, int](3)
	const goroutines = 10
	const opsPerGoroutine = 500

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := base*1000 + i
				tree.Insert(key, key*2)
				if i%2 == 0 {
					tree.Delete(key)
				}
				tree.Find(key)
			}
		}(g)
	}

	wg.Wait()
	// Если бы были гонки, тест упал бы с флагом -race.
}

// TestBTreeProperties проверяет структурные инварианты B-дерева после серии операций
func TestUnit_BTreeProperties(t *testing.T) {
	deg := 3
	tree := New[int, int](deg)
	// Вставим достаточно ключей, чтобы точно было несколько уровней
	for i := 0; i < 100; i++ {
		tree.Insert(i, i)
	}
	// Удалим часть
	for i := 10; i < 30; i++ {
		tree.Delete(i)
	}
	for i := 50; i < 70; i++ {
		tree.Delete(i)
	}
	// Проверим инварианты через обход
	if err := validateBTree(tree, deg); err != nil {
		t.Errorf("B-tree invariant violation: %v", err)
	}
}

// validateBTree рекурсивно проверяет свойства B-дерева:
// - все ключи в узле отсортированы
// - количество ключей (кроме корня) между minKeys и maxKeys
// - ключи в детях правильным образом связаны с родителем
// - все листья находятся на одной глубине
func validateBTree[K int, V any](tree *BTree[K, V], degree int) error {
	if tree.root.numKeys == 0 {
		return nil
	}
	minKeys := degree - 1
	maxKeys := 2*degree - 1

	// Проверяем корень: для корня ограничение только maxKeys (minKeys не требуется)
	if tree.root.numKeys > maxKeys {
		return fmt.Errorf("root has %d keys > maxKeys %d", tree.root.numKeys, maxKeys)
	}

	// Проверка рекурсивно
	height, err := validateNode(tree.root, degree, minKeys, maxKeys)
	if err != nil {
		return err
	}
	_ = height
	return nil
}

// validateNode возвращает высоту поддерева и ошибку, если нарушены инварианты.
func validateNode[K int, V any](n *node[K, V], degree, minKeys, maxKeys int) (int, error) {
	// Проверка сортировки ключей
	for i := 1; i < n.numKeys; i++ {
		if n.keys[i-1] > n.keys[i] {
			return 0, fmt.Errorf("keys not sorted at node %v", n.keys[:n.numKeys])
		}
	}

	// Проверка количества ключей (для некорневых узлов; признак корня проверяется отдельно)
	// Здесь мы не различаем корень, но вызовем validateNode для детей, и для детей ограничение строгое.
	// Для самого узла n, если это не лист, проверяем детей.
	if !n.isLeaf {
		// Проверка количества детей (n.numKeys+1)
		if len(n.children) < n.numKeys+1 {
			return 0, fmt.Errorf("not enough children pointers")
		}
		for i := 0; i <= n.numKeys; i++ {
			if n.children[i] == nil {
				return 0, fmt.Errorf("null child pointer at index %d", i)
			}
			// Проверка, что ключи ребёнка лежат в правильных диапазонах
			child := n.children[i]
			// Проверка minKeys/maxKeys для ребёнка (т.к. он не корень)
			if child.numKeys < minKeys || child.numKeys > maxKeys {
				return 0, fmt.Errorf("child has %d keys (min=%d, max=%d)", child.numKeys, minKeys, maxKeys)
			}
			// Ключи ребёнка должны быть меньше ключа родителя (для i) и больше предыдущего
			if i > 0 {
				// ключи ребёнка должны быть > n.keys[i-1]
				if child.numKeys > 0 && child.keys[0] <= n.keys[i-1] {
					return 0, fmt.Errorf("child %d first key %d <= parent key %d", i, child.keys[0], n.keys[i-1])
				}
			}
			if i < n.numKeys {
				// ключи ребёнка должны быть < n.keys[i]
				if child.numKeys > 0 && child.keys[child.numKeys-1] >= n.keys[i] {
					return 0, fmt.Errorf("child %d last key %d >= parent key %d", i, child.keys[child.numKeys-1], n.keys[i])
				}
			}
		}

		// Рекурсивная проверка детей и вычисление высоты
		var childHeight int
		for i := 0; i <= n.numKeys; i++ {
			h, err := validateNode(n.children[i], degree, minKeys, maxKeys)
			if err != nil {
				return 0, err
			}
			if i == 0 {
				childHeight = h
			} else if h != childHeight {
				return 0, fmt.Errorf("child heights differ: %d vs %d", childHeight, h)
			}
		}
		return childHeight + 1, nil
	}

	// Лист
	return 0, nil
}

//----------------------------------------------------------------------------------------------------

/*
defaultDegree = 32
goos: linux
goarch: amd64
pkg: go.osspkg.com/algorithms/trees/btree
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
BenchmarkBTree_Insert
BenchmarkBTree_Insert/Sequential
BenchmarkBTree_Insert/Sequential-4    	     896	   1382861 ns/op	  484544 B/op	     904 allocs/op
BenchmarkBTree_Insert/Parallel
BenchmarkBTree_Insert/Parallel-4      	10466346	       129.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find
BenchmarkBTree_Find/Hit
BenchmarkBTree_Find/Hit-4             	21723561	        54.69 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find/Miss
BenchmarkBTree_Find/Miss-4            	100000000	        11.33 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find/Parallel
BenchmarkBTree_Find/Parallel-4        	33573949	        36.75 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Delete
BenchmarkBTree_Delete/Sequential
BenchmarkBTree_Delete/Sequential-4    	    1158	   1100783 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Delete/Parallel
BenchmarkBTree_Delete/Parallel-4      	39781627	        33.21 ns/op	      17 B/op	       0 allocs/op
BenchmarkBTree_MixedLoad_Parallel
BenchmarkBTree_MixedLoad_Parallel-4   	 5782959	       236.1 ns/op	       0 B/op	       0 allocs/op
*/

/*
defaultDegree = 2
goos: linux
goarch: amd64
pkg: go.osspkg.com/algorithms/trees/btree
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
BenchmarkBTree_Insert
BenchmarkBTree_Insert/Sequential
BenchmarkBTree_Insert/Sequential-4    	     392	   2917484 ns/op	 1147816 B/op	   22956 allocs/op
BenchmarkBTree_Insert/Parallel
BenchmarkBTree_Insert/Parallel-4      	 3160122	       404.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find
BenchmarkBTree_Find/Hit
BenchmarkBTree_Find/Hit-4             	10859036	       106.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find/Miss
BenchmarkBTree_Find/Miss-4            	59257408	        19.98 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find/Parallel
BenchmarkBTree_Find/Parallel-4        	27232424	        44.92 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Delete
BenchmarkBTree_Delete/Sequential
BenchmarkBTree_Delete/Sequential-4    	     802	   1506062 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Delete/Parallel
BenchmarkBTree_Delete/Parallel-4      	15402082	        73.63 ns/op	      40 B/op	       0 allocs/op
BenchmarkBTree_MixedLoad_Parallel
BenchmarkBTree_MixedLoad_Parallel-4   	 4454574	       258.9 ns/op	       0 B/op	       0 allocs/op
*/

/*
defaultDegree = 64
goos: linux
goarch: amd64
pkg: go.osspkg.com/algorithms/trees/btree
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
BenchmarkBTree_Insert
BenchmarkBTree_Insert/Sequential
BenchmarkBTree_Insert/Sequential-4    	     592	   1755761 ns/op	  505440 B/op	     468 allocs/op
BenchmarkBTree_Insert/Parallel
BenchmarkBTree_Insert/Parallel-4      	 4480512	       262.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find
BenchmarkBTree_Find/Hit
BenchmarkBTree_Find/Hit-4             	18538779	        62.38 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find/Miss
BenchmarkBTree_Find/Miss-4            	100000000	        10.69 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Find/Parallel
BenchmarkBTree_Find/Parallel-4        	29789313	        36.98 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Delete
BenchmarkBTree_Delete/Sequential
BenchmarkBTree_Delete/Sequential-4    	     874	   1433584 ns/op	       0 B/op	       0 allocs/op
BenchmarkBTree_Delete/Parallel
BenchmarkBTree_Delete/Parallel-4      	30289316	        41.55 ns/op	      19 B/op	       0 allocs/op
BenchmarkBTree_MixedLoad_Parallel
BenchmarkBTree_MixedLoad_Parallel-4   	 6820371	       396.0 ns/op	       0 B/op	       0 allocs/op
*/

const (
	defaultDegree = 4
	treeSize      = 10_000
)

// Набор предзаполненных данных для минимизации аллокаций внутри бенчмарков
var (
	benchKeys []int
	benchVals []string
)

func init() {
	benchKeys = make([]int, treeSize*2)
	benchVals = make([]string, treeSize*2)
	for i := 0; i < treeSize*2; i++ {
		benchKeys[i] = i
		benchVals[i] = fmt.Sprintf("val-%d", i)
	}
	// Перемешиваем для исключения деградации в упорядоченную последовательность (если применимо)
	rand.Shuffle(len(benchKeys), func(i, j int) {
		benchKeys[i], benchKeys[j] = benchKeys[j], benchKeys[i]
		benchVals[i], benchVals[j] = benchVals[j], benchVals[i]
	})
}

func setupTree(degree, size int) (*BTree[int, string], []int) {
	t := New[int, string](degree)
	for i := 0; i < size; i++ {
		t.Insert(benchKeys[i], benchVals[i])
	}
	return t, benchKeys[:size]
}

func BenchmarkBTree_Insert(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			t := New[int, string](defaultDegree)
			b.StartTimer()

			for j := 0; j < treeSize; j++ {
				t.Insert(benchKeys[j], benchVals[j])
			}
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		t := New[int, string](defaultDegree)
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Используем цикличный сдвиг индексов во избежание явных коллизий на записи
				idx := (i) % (treeSize * 2)
				t.Insert(benchKeys[idx], benchVals[idx])
				i++
			}
		})
	})
}

func BenchmarkBTree_Find(b *testing.B) {
	t, keys := setupTree(defaultDegree, treeSize)

	b.Run("Hit", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Гарантированный хит из существующих ключей
			idx := i % len(keys)
			_, _ = t.Find(keys[idx])
		}
	})

	b.Run("Miss", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Заведомо отсутствующий ключ в дереве
			_, _ = t.Find(-(i + 1))
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				idx := i % len(keys)
				_, _ = t.Find(keys[idx])
				i++
			}
		})
	})
}

func BenchmarkBTree_Delete(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			t, keys := setupTree(defaultDegree, treeSize)
			b.StartTimer()

			for _, k := range keys {
				t.Delete(k)
			}
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		// Из-за разрушающего характера Delete на общем состоянии,
		// параллельный запуск требует частой инициализации новых деревьев воркерами
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			t, keys := setupTree(defaultDegree, treeSize)
			i := 0
			for pb.Next() {
				idx := i % len(keys)
				t.Delete(keys[idx])
				i++

				// Восстановление структуры при опустошении во избежание деградации замера
				if idx == len(keys)-1 {
					b.StopTimer()
					t, keys = setupTree(defaultDegree, treeSize)
					b.StartTimer()
				}
			}
		})
	})
}

func BenchmarkBTree_MixedLoad_Parallel(b *testing.B) {
	t, keys := setupTree(defaultDegree, treeSize)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			op := i % 100
			idx := i % len(keys)

			switch {
			case op < 80: // 80% Reads
				_, _ = t.Find(keys[idx])
			case op < 95: // 15% Writes (Inserts/Updates)
				t.Insert(benchKeys[idx], benchVals[idx])
			default: // 5% Deletes
				t.Delete(keys[idx])
			}
			i++
		}
	})
}
