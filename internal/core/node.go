package core

import (
	"time"
)

// Node is basic type of in-game server
type Node struct {
	// mu                 sync.RWMutex  // for future optimization (Per-node locking)
	id                 string        // stands for IP-address
	isHacked           bool          // server state for hacking process
	passwordComplexity time.Duration // stands for password complexity, expressed in time for cracking process
}

// NewNode is a constructor for a new instance of Node
func NewNode(ID string, passwdTime time.Duration) *Node {
	return &Node{
		id:                 ID,
		passwordComplexity: passwdTime,
	}
}

// IsCracked checks node condition
func (n *Node) IsCracked() bool {
	return n.isHacked
}

// GetPassComplexity returns node's password complexity (base cracking time)
func (n *Node) GetPassComplexity() time.Duration {
	return n.passwordComplexity
}
