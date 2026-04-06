package core

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func setupEngine() *Engine {
	e := NewEngine()
	e.AddNode("192.168.0.1", 10*time.Minute)
	e.AddNode("10.0.0.1", 10*time.Millisecond)
	e.AddNode("1.1.1.1", 10*time.Second)
	return e
}

func TestJohnTheRipper_Success(t *testing.T) {
	e := setupEngine()
	nodeID := "10.0.0.1"

	err := e.JohnTheRipper(nodeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	cracked, err := e.IsNodeCracked(nodeID)
	if err != nil {
		t.Fatalf("failed to get node status: %v", err)
	}
	if !cracked {
		t.Error("node should be cracked after attack")
	}
}

func TestJohnTheRipper_NodeNotExists(t *testing.T) {
	e := setupEngine()
	err := e.JohnTheRipper("8.8.8.8")
	if !errors.Is(err, ErrNodeNotExists) {
		t.Errorf("expected %v, got %v", ErrNodeNotExists, err)
	}
}

func TestJohnTheRipper_AlreadyCracked(t *testing.T) {
	e := setupEngine()
	nodeID := "10.0.0.1"

	err := e.JohnTheRipper(nodeID)
	if err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	err = e.JohnTheRipper(nodeID)
	if !errors.Is(err, ErrNodeIsCracked) {
		t.Errorf("expected %v, got %v", ErrNodeIsCracked, err)
	}
}

func TestJohnTheRipper_ConcurrentAttacksOnSameNode(t *testing.T) {
	e := setupEngine()
	nodeID := "10.0.0.1"

	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	wg.Add(10)
	for range 10 {
		go func() {
			defer wg.Done()
			errCh <- e.JohnTheRipper(nodeID)
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil && !errors.Is(err, ErrNodeIsCracked) {
			t.Errorf("unexpected error during concurrent attack: %v", err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	cracked, err := e.IsNodeCracked(nodeID)
	if err != nil {
		t.Fatalf("failed to check node: %v", err)
	}
	if !cracked {
		t.Error("node should be cracked after concurrent attacks")
	}
}

func TestSnapshotConsistency(t *testing.T) {
	e := setupEngine()

	snapBefore := e.Snapshot()
	if len(snapBefore) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(snapBefore))
	}

	err := e.JohnTheRipper("10.0.0.1")
	if err != nil {
		t.Fatalf("failed to start attack: %v", err)
	}

	snapDuring := e.Snapshot()
	time.Sleep(10 * time.Millisecond)
	snapAfter := e.Snapshot()

	foundDuring := false
	foundAfter := false
	for _, n := range snapDuring {
		if n.id == "10.0.0.1" && n.IsCracked() {
			foundDuring = true
		}
	}
	for _, n := range snapAfter {
		if n.id == "10.0.0.1" && n.IsCracked() {
			foundAfter = true
		}
	}

	if foundDuring {
		t.Error("snapshot during attack should not show cracked node")
	}
	if !foundAfter {
		t.Error("snapshot after attack should show cracked node")
	}
}

func TestNetCat(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  string
		wantErr bool
		errBody error
	}{
		{
			name:    "connect successful",
			nodeID:  "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "node not found",
			nodeID:  "404",
			wantErr: true,
			errBody: ErrNodeNotExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupEngine()

			if err := e.NetCat(tt.nodeID); err != nil {
				if !tt.wantErr {
					t.Errorf("Got unexpected error: %v", err)
					return
				}
				if err != tt.errBody {
					t.Errorf("Want err: %v, got: %v", tt.errBody, err)
					return
				}
				return
			}

			if tt.wantErr {
				t.Errorf("Expected error: %v, got none", tt.errBody)
			}
		})
	}
}
