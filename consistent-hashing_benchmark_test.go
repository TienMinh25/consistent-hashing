package consistenthashing

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkGetNode(b *testing.B) {
	hash := NewConsistentHash(DefaultHashFunction, 100)
	for i := 0; i < 10; i++ {
		hash.AddNode(Node{ID: fmt.Sprintf("node-%d", i), Weight: 1})
	}

	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hash.GetNode(keys[i%len(keys)])
	}
}

func BenchmarkGetNode_Parallel(b *testing.B) {
	hash := NewConsistentHash(DefaultHashFunction, 100)
	for i := 0; i < 10; i++ {
		hash.AddNode(Node{ID: fmt.Sprintf("node-%d", i), Weight: 1})
	}

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(rand.Int63()))
		for pb.Next() {
			key := fmt.Sprintf("key-%d", r.Intn(100000))
			_, _ = hash.GetNode(key)
		}
	})
}

func BenchmarkAddNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hash := NewConsistentHash(DefaultHashFunction, 100)
		for j := 0; j < 10; j++ {
			hash.AddNode(Node{ID: fmt.Sprintf("seed-%d", j), Weight: 1})
		}
		b.StartTimer()

		hash.AddNode(Node{ID: fmt.Sprintf("new-node-%d", i), Weight: 1})
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hash := NewConsistentHash(DefaultHashFunction, 100)
		nodes := make([]Node, 10)
		for j := 0; j < 10; j++ {
			nodes[j] = Node{ID: fmt.Sprintf("node-%d", j), Weight: 1}
			hash.AddNode(nodes[j])
		}
		b.StartTimer()

		hash.RemoveNode(nodes[0])
	}
}

func BenchmarkAddNode_ScalingRingSize(b *testing.B) {
	for _, existingNodes := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("existing=%d", existingNodes), func(b *testing.B) {
			hash := NewConsistentHash(DefaultHashFunction, 100)
			for j := 0; j < existingNodes; j++ {
				hash.AddNode(Node{ID: fmt.Sprintf("seed-%d", j), Weight: 1})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nodeID := fmt.Sprintf("new-node-%d", i)
				hash.AddNode(Node{ID: nodeID, Weight: 1})

				b.StopTimer()
				hash.RemoveNode(Node{ID: nodeID})
				b.StartTimer()
			}
		})
	}
}