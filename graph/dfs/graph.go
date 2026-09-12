/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dfs

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrCycleDetected = errors.New("cycle detected: graph is not a DAG")
	ErrNodeNotFound  = errors.New("node not found")
	ErrNodeKeyExist  = errors.New("node key exist")
)

const (
	stateVisiting uint8 = 1
	stateDone     uint8 = 2
)

type Graph[K comparable] struct {
	mu        sync.RWMutex
	nodes     map[K]struct{}
	adjacency map[K]map[K]struct{}
}

func NewGraph[K comparable]() *Graph[K] {
	return &Graph[K]{
		nodes:     make(map[K]struct{}, 10),
		adjacency: make(map[K]map[K]struct{}, 10),
	}
}

func (g *Graph[K]) AddNode(key K) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.nodes[key]; ok {
		return fmt.Errorf("%w: %v", ErrNodeKeyExist, key)
	}

	g.nodes[key] = struct{}{}
	if _, exists := g.adjacency[key]; !exists {
		g.adjacency[key] = make(map[K]struct{})
	}

	return nil
}

func (g *Graph[K]) AddEdge(from, to K) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[from]; !exists {
		return fmt.Errorf("%w: %v", ErrNodeNotFound, from)
	}
	if _, exists := g.nodes[to]; !exists {
		return fmt.Errorf("%w: %v", ErrNodeNotFound, to)
	}

	g.adjacency[from][to] = struct{}{}
	return nil
}

func (g *Graph[K]) TopologicalSort() ([]K, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[K]uint8, len(g.nodes))
	order := make([]K, 0, len(g.nodes))

	var dfs func(node K) error
	dfs = func(node K) error {
		switch visited[node] {
		case stateVisiting:
			return ErrCycleDetected
		case stateDone:
			return nil
		}

		visited[node] = stateVisiting

		for neighbor := range g.adjacency[node] {
			if err := dfs(neighbor); err != nil {
				return err
			}
		}

		visited[node] = stateDone
		order = append(order, node)

		return nil
	}

	for node := range g.nodes {
		if _, exists := visited[node]; !exists {
			if err := dfs(node); err != nil {
				return nil, err
			}
		}
	}

	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	return order, nil
}
