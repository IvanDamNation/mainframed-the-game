package core

import "sync"

// World is basic type of in-game world (network)
type World struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

// NewWorld is a constructor for in-game world
func NewWorld() *World {
	return &World{
		nodes: make(map[string]*Node),
	}
}
