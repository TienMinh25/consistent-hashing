package consistenthashing

import (
	"testing"
)

func makeHashFunc(mapping map[string]uint32) HashFunc {
	return func(key string) uint32 {
		return mapping[key]
	}
}

func TestAddNode(t *testing.T) {
	t.Run("add node successfully", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1": 100,
		}))

		if err := hash.AddNode(Node{
			ID: "node-1",
		}); err != nil {
			t.Fatalf("expected no error when add one node, got %v", err.Error())
		}
	})

	t.Run("add one node twice -> should return error in turn 2", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1": 100,
		}))

		if err := hash.AddNode(Node{ID: "node-1"}); err != nil {
			t.Fatalf("expected no error when add one node, got %v", err.Error())
		}

		if err := hash.AddNode(Node{ID: "node-1"}); err == nil {
			t.Fatalf("expected error when add twice twice, got %v", err.Error())
		}
	})
}

func TestRemoveNode(t *testing.T) {
	t.Run("remove node successfully", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1": 100,
		}))

		hash.AddNode(Node{ID: "node-1"})

		if err := hash.RemoveNode(Node{ID: "node-1"}); err != nil {
			t.Fatalf("expected no error when remove one node, got %v", err.Error())
		}
	})

	t.Run("remove one node twice -> successfully", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1": 100,
		}))

		hash.AddNode(Node{ID: "node-1"})

		if err := hash.RemoveNode(Node{ID: "node-1"}); err != nil {
			t.Fatalf("expected no error when remove one node, got %v", err.Error())
		}

		if err := hash.RemoveNode(Node{ID: "node-1"}); err != nil {
			t.Fatalf("expected no error when remove one node, got %v", err.Error())
		}
	})
}

func TestGetNode(t *testing.T) {
	t.Run("get node successfully", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1": 100,
			"node-2": 200,
			"key1":   115,
		}))

		hash.AddNode(Node{ID: "node-1"})
		hash.AddNode(Node{ID: "node-2"})

		node, err := hash.GetNode("key1")
		if err != nil {
			t.Fatalf("expected no error when get one node, got %v", err.Error())
		}

		if node.ID != "node-2" {
			t.Fatalf("expected node to be node-1, got %v", node)
		}
	})

	t.Run("get node while ring is empty -> error", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{}))

		node, err := hash.GetNode("node-1")
		if err == nil {
			t.Fatalf("expected error when get one node, got %v", err.Error())
		}

		if node.ID != "" {
			t.Fatalf("expected node to be empty, got %v", node)
		}
	})

	t.Run("same key -> same node", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1":   100,
			"node-2":   200,
			"node-3":   300,
			"user-123": 1121,
		}))
		hash.AddNode(Node{ID: "node-1"})
		hash.AddNode(Node{ID: "node-2"})
		hash.AddNode(Node{ID: "node-3"})

		n1, _ := hash.GetNode("user-123")
		n2, _ := hash.GetNode("user-123")

		if n1 != n2 {
			t.Fatalf("expected same key always get the same node")
		}
	})

	t.Run("key hash greater than all server hash value -> return node has smallest hash value", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint32{
			"node-1":   100,
			"node-2":   200,
			"node-3":   300,
			"user-123": 1121,
		}))

		hash.AddNode(Node{ID: "node-2"})
		hash.AddNode(Node{ID: "node-3"})
		hash.AddNode(Node{ID: "node-1"})

		node, _ := hash.GetNode("user-123")

		if node.ID != "node-1" {
			t.Fatalf("expected node to be node-1, got %v", node)
		}
	})
}

func TestConsistentHash_AddNodeAffectsOnlySubsetOfKeys(t *testing.T) {
	mapping := map[string]uint32{
		"node-1": 100,
		"node-2": 200,
		"node-3": 300,
		"node-4": 400,
		"key-1":  55,
		"key-2":  103,
		"key-3":  250,
		"key-4":  350,
	}
	hash := NewConsistentHash(makeHashFunc(mapping))

	hash.AddNode(Node{ID: "node-1"})
	hash.AddNode(Node{ID: "node-3"})
	hash.AddNode(Node{ID: "node-2"})

	keys := []string{
		"key-1",
		"key-2",
		"key-3",
		"key-4",
	}

	before := make(map[string]string)
	for _, key := range keys {
		node, _ := hash.GetNode(key)
		before[key] = node.ID
	}

	hash.AddNode(Node{ID: "node-4"})
	after := make(map[string]string)
	for _, key := range keys {
		node, _ := hash.GetNode(key)
		after[key] = node.ID
	}

	moved := 0

	for key := range before {
		if after[key] != before[key] {
			moved++
		}
	}

	if moved == 0 {
		t.Fatal("expected some keys remapped")
	}

	if moved >= len(keys) {
		t.Fatal("expected only subset of keys remapped")
	}
}

func TestConsistentHash_RemoveNodeAffectsOnlySubsetOfKeys(t *testing.T) {
	mapping := map[string]uint32{
		"node-1": 100,
		"node-2": 200,
		"node-3": 300,
		"node-4": 400,
		"key-1":  55,
		"key-2":  103,
		"key-3":  250,
		"key-4":  350,
	}
	hash := NewConsistentHash(makeHashFunc(mapping))

	hash.AddNode(Node{ID: "node-1"})
	hash.AddNode(Node{ID: "node-2"})
	hash.AddNode(Node{ID: "node-3"})
	hash.AddNode(Node{ID: "node-4"})

	keys := []string{
		"key-1",
		"key-2",
		"key-3",
		"key-4",
	}

	before := make(map[string]string)
	for _, key := range keys {
		node, _ := hash.GetNode(key)
		before[key] = node.ID
	}

	hash.RemoveNode(Node{ID: "node-4"})
	after := make(map[string]string)
	for _, key := range keys {
		node, _ := hash.GetNode(key)
		after[key] = node.ID
	}

	moved := 0
	for key := range before {
		if after[key] != before[key] {
			moved++
		}
	}

	if moved == 0 {
		t.Fatal("expected some keys remapped")
	}

	if moved >= len(keys) {
		t.Fatal("expected only subset of keys remapped")
	}
}

func TestConsistentHash_VirtualNodeImproveDistributionEvenly(t *testing.T) {

}
