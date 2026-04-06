package core

import (
	"errors"
	"time"
)

var (
	ErrNodeIsCracked = errors.New("node is already cracked")
	ErrNodeNotExists = errors.New("node not exists")
)

// Simulation engine. Main object for UI
type Engine struct {
	world *World
}

// NewEngine constructs new simulation engine
func NewEngine() *Engine {
	return &Engine{
		world: NewWorld(),
	}
}

// AddNode adds new Node instance to game world
func (e *Engine) AddNode(id string, crackTime time.Duration) {
	node := NewNode(id, crackTime)

	e.world.mu.Lock()
	defer e.world.mu.Unlock()

	e.world.nodes[id] = node
}

// Snapshot gets state for all servers (Nodes) (for rendering)
func (e *Engine) Snapshot() []Node {
	e.world.mu.RLock()
	defer e.world.mu.RUnlock()

	// TODO: sync.Pool for nodeList ?
	nodeList := make([]Node, 0, len(e.world.nodes))
	for _, node := range e.world.nodes {
		nodeList = append(nodeList, *node)
	}

	return nodeList
}

// NetCat is a player utility action.
// It tries connect to Node by ID
func (e *Engine) NetCat(nodeID string) error {
	e.world.mu.RLock()
	defer e.world.mu.RUnlock()
	if _, exists := e.world.nodes[nodeID]; !exists {
		return ErrNodeNotExists
	}
	// TODO: establish connection mechanic
	return nil
}

// JohnTheRipper is a player attack action.
// It cracks Node in passwordDifficulty (time based)
func (e *Engine) JohnTheRipper(nodeID string) error {
	var node *Node
	var exists bool

	e.world.mu.RLock()
	defer e.world.mu.RUnlock()

	if node, exists = e.world.nodes[nodeID]; !exists {
		return ErrNodeNotExists
	}

	// TODO: connection must be established for cracking
	if cracked := node.IsCracked(); cracked {
		return ErrNodeIsCracked
	}

	passComplexity := node.GetPassComplexity()
	go e.passCrackingProc(nodeID, passComplexity)
	return nil
}

// NOTE: Func for future optimization (Per-node locking). Do not forget change e.world.nodes[id] to getNode(id)
// getNode is a method for checking node existing and return exact node by it's id
// func (e *Engine) getNode(id string) (*Node, bool) {
// 	e.world.mu.RLock()
// 	defer e.world.mu.RUnlock()
// 	node, exists := e.world.nodes[id]
// 	return node, exists
// }

// passCrackingProc is a async worker for JohnTheRipper
func (e *Engine) passCrackingProc(id string, complexity time.Duration) {
	<-time.After(complexity)

	e.world.mu.Lock()
	defer e.world.mu.Unlock()

	node, exists := e.world.nodes[id]
	if !exists {
		// TODO: log node err
		return
	}

	if !node.IsCracked() {
		node.isHacked = true
	}

	// TODO: log node event
}

// NOTE: Race condition resolver. Till Per-Node lock
// IsNodeCracked returns true if node is already hacked, false otherwise.
// Returns error if node does not exist.
func (e *Engine) IsNodeCracked(nodeID string) (bool, error) {
	e.world.mu.RLock()
	defer e.world.mu.RUnlock()

	node, exists := e.world.nodes[nodeID]
	if !exists {
		return false, ErrNodeNotExists
	}
	return node.isHacked, nil
}
