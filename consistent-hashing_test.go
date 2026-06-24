package consistenthashing

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

func makeHashFunc(mapping map[string]uint64) HashFunc {
	return func(key string) uint64 {
		return mapping[key]
	}
}

func TestAddNode(t *testing.T) {
	t.Run("add node successfully", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1": 100,
		}), 100)

		if err := hash.AddNode(Node{
			ID: "node-1",
		}); err != nil {
			t.Fatalf("expected no error when add one node, got %v", err.Error())
		}
	})

	t.Run("add one node twice -> should return error in turn 2", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1": 100,
		}), 100)

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
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1": 100,
		}), 100)

		hash.AddNode(Node{ID: "node-1"})

		if err := hash.RemoveNode(Node{ID: "node-1"}); err != nil {
			t.Fatalf("expected no error when remove one node, got %v", err.Error())
		}
	})

	t.Run("remove one node twice -> error", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1": 100,
		}), 100)

		hash.AddNode(Node{ID: "node-1"})

		if err := hash.RemoveNode(Node{ID: "node-1"}); err != nil {
			t.Fatalf("expected no error when remove one node, got %v", err.Error())
		}

		if err := hash.RemoveNode(Node{ID: "node-1"}); err == nil {
			t.Fatalf("expected error when remove one node twice, got %v", err.Error())
		}
	})
}

func TestGetNode(t *testing.T) {
	t.Run("get node successfully", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1": 100,
			"node-2": 200,
			"key1":   115,
		}), 100)

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
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{}), 100)

		node, err := hash.GetNode("node-1")
		if err == nil {
			t.Fatalf("expected error when get one node, got %v", err.Error())
		}

		if node.ID != "" {
			t.Fatalf("expected node to be empty, got %v", node)
		}
	})

	t.Run("same key -> same node", func(t *testing.T) {
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1":   100,
			"node-2":   200,
			"node-3":   300,
			"user-123": 1121,
		}), 100)
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
		hash := NewConsistentHash(makeHashFunc(map[string]uint64{
			"node-1":   100,
			"node-2":   200,
			"node-3":   300,
			"user-123": 1121,
		}), 100)

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
	mapping := map[string]uint64{
		"node-1": 100,
		"node-2": 200,
		"node-3": 300,
		"node-4": 400,
		"key-1":  55,
		"key-2":  103,
		"key-3":  250,
		"key-4":  350,
	}
	hash := NewConsistentHash(makeHashFunc(mapping), 100)

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
	mapping := map[string]uint64{
		"node-1": 100,
		"node-2": 200,
		"node-3": 300,
		"node-4": 400,
		"key-1":  55,
		"key-2":  103,
		"key-3":  250,
		"key-4":  350,
	}
	hash := NewConsistentHash(makeHashFunc(mapping), 100)

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
	const (
		numNodes        = 10
		numVirtualNodes = 100
		numKeys         = 10000
	)
	hash := NewConsistentHash(DefaultHashFunction, 100)

	for idx := 0; idx < numNodes; idx++ {
		hash.AddNode(Node{ID: fmt.Sprintf("node-%d", idx), Weight: 1})
	}

	counts := make(map[string]int)
	for idx := 0; idx < numKeys; idx++ {
		key := fmt.Sprintf("key-%d", idx)
		node, _ := hash.GetNode(key)
		counts[node.ID]++
	}

	mean := float64(numKeys) / float64(numNodes) // the average number of keys per node

	var sumSqDiff float64
	for key, count := range counts {
		fmt.Println(key, count)
		diff := float64(count) - mean // the deviation of each node from the desired average value
		sumSqDiff += diff * diff      // the sum of the squares of the deviations
	}
	stdDev := math.Sqrt(sumSqDiff / float64(numNodes)) // the standard deviation of the number of keys per node
	cv := stdDev / mean                                // percentage of standard deviation

	t.Logf("mean=%v stdDev=%v cv=%.2f%%", mean, stdDev, cv*100)
	if cv > 0.15 {
		t.Fatalf("expected coefficient of variation <= 15%%, got %.2f%%", cv*100)
	}
}

func TestConsistentHash_WeightedDistribution(t *testing.T) {
	hash := NewConsistentHash(DefaultHashFunction, 100) // 100 is base virtual node

	hash.AddNode(Node{ID: "node-1", Weight: 1})
	hash.AddNode(Node{ID: "node-3", Weight: 3})
	hash.AddNode(Node{ID: "node-2", Weight: 2})

	counts := make(map[string]int)
	const numKeys = 60000
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key-%d", i)
		node, _ := hash.GetNode(key)
		counts[node.ID]++
	}

	t.Logf("counts=%v", counts)

	totalWeight := 6 // sum of weights: 1 + 3 + 2
	expectedDistribution := map[string]float64{
		"node-1": numKeys * 1.0 / float64(totalWeight),
		"node-2": numKeys * 2.0 / float64(totalWeight),
		"node-3": numKeys * 3.0 / float64(totalWeight),
	}

	for id, expected := range expectedDistribution {
		actual := float64(counts[id])
		ratio := actual / expected
		// margin % is 15% (0.85 to 1.15)
		if ratio < 0.85 || ratio > 1.15 {
			t.Fatalf("%s: expected ~%.0f keys, got %d (ratio=%.2f)", id, expected, counts[id], ratio)
		}
	}
}

func TestConsistentHash_ConcurrentAccess(t *testing.T) {
	hash := NewConsistentHash(DefaultHashFunction, 100)

	for i := 0; i < 5; i++ {
		hash.AddNode(Node{ID: fmt.Sprintf("seed-node-%d", i), Weight: 1})
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				_, _ = hash.GetNode(key)
			}
		}(i)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			nodeID := fmt.Sprintf("dynamic-node-%d", id)
			hash.AddNode(Node{ID: nodeID, Weight: 1})
			time.Sleep(time.Millisecond)
			hash.RemoveNode(Node{ID: nodeID})
		}(i)
	}

	wg.Wait()
}
