# consistent-hashing

A from-scratch, TDD-driven implementation of consistent hashing in Go, with virtual nodes, weighted nodes, and concurrency safety.

This project was built incrementally to understand consistent hashing at a level deeper than "draw a ring and explain it" — each feature (virtual nodes, weighted distribution, thread safety) was driven by a failing test first, then an implementation, then a benchmark to validate the performance characteristics.

## Features

- **Consistent hashing core** — keys are mapped to nodes via a sorted hash ring; adding or removing a node only reshuffles a bounded slice of keys, not the whole keyspace.
- **Virtual nodes** — each physical node is represented by many points on the ring, which smooths out the uneven load distribution that plain consistent hashing suffers from when there are few nodes.
- **Weighted nodes** — nodes can be assigned a `Weight`, which scales their virtual node count proportionally. A node with `Weight: 2` gets twice the ring real estate (and therefore roughly twice the traffic) of a node with `Weight: 1`.
- **Thread-safe** — all operations are safe for concurrent use via `sync.RWMutex`, tuned for a read-heavy / write-rare access pattern (`GetNode` is called far more often than `AddNode`/`RemoveNode`).
- **MD5-based default hashing** — chosen after benchmarking showed CRC32 produces a noticeably less uniform distribution than MD5 at this scale.

## Installation

```bash
go get github.com/TienMinh25/consistent-hashing
```

## Usage

```go
package main

import (
	"fmt"

	consistenthashing "github.com/TienMinh25/consistent-hashing"
)

func main() {
	// baseVirtualNodes controls how many ring points a Weight: 1 node gets.
	// Higher values -> more even distribution, more memory.
	hash := consistenthashing.NewConsistentHash(consistenthashing.DefaultHashFunction, 100)

	hash.AddNode(consistenthashing.Node{ID: "node-1", Weight: 1})
	hash.AddNode(consistenthashing.Node{ID: "node-2", Weight: 1})
	hash.AddNode(consistenthashing.Node{ID: "node-3", Weight: 2}) // gets ~2x the keys

	node, err := hash.GetNode("some-key")
	if err != nil {
		panic(err)
	}
	fmt.Printf("key routed to: %s\n", node.ID)

	hash.RemoveNode(consistenthashing.Node{ID: "node-2"})
}
```

## API

```go
type ConsistentHash interface {
	GetNode(keyStr string) (Node, error)
	AddNode(node Node) error
	RemoveNode(node Node) error
}

type Node struct {
	ID     string
	Weight int
	Hash   uint64
}

func NewConsistentHash(hashFunc HashFunc, baseVirtualNodes int) ConsistentHash
```

- `GetNode` — returns the node responsible for a key. Read-locked; safe for high-concurrency lookups.
- `AddNode` — adds a node and its virtual nodes (`baseVirtualNodes * node.Weight` of them) to the ring. Returns an error if the node ID already exists.
- `RemoveNode` — removes a node and all of its virtual nodes from the ring. Returns an error if the node doesn't exist.

## Benchmarks

Run with:

```bash
go test -bench=. -benchmem -run=^$ ./...
```

```
goos: linux
goarch: amd64
pkg: github.com/TienMinh25/consistent-hashing
cpu: 11th Gen Intel(R) Core(TM) i5-11400 @ 2.60GHz
BenchmarkGetNode-12                      8973247               132.1 ns/op             0 B/op          0 allocs/op
BenchmarkGetNode_Parallel-12            21347702                55.66 ns/op           23 B/op          1 allocs/op
BenchmarkAddNode-12                        49593             27726 ns/op           48482 B/op        178 allocs/op
BenchmarkRemoveNode-12                     31150             38959 ns/op           39464 B/op        200 allocs/op
BenchmarkAddNode_ScalingRingSize/existing=10-12                    44798             28287 ns/op           48443 B/op        177 allocs/op
BenchmarkAddNode_ScalingRingSize/existing=100-12                   16762             72295 ns/op          334457 B/op        143 allocs/op
BenchmarkAddNode_ScalingRingSize/existing=1000-12                   2772            488564 ns/op         3242933 B/op        110 allocs/op
PASS
ok      github.com/TienMinh25/consistent-hashing        54.125s
```

## Testing

```bash
go test -v ./...              # correctness tests
go test -race ./...           # concurrency safety
go test -bench=. -benchmem -run=^$ ./...   # performance
```

The test suite includes:
- Correctness tests for routing, add, and remove.
- A statistical distribution test (`TestConsistentHash_VirtualNodeImproveDistributionEvenly`) that verifies the coefficient of variation across nodes stays within a defined bound when virtual nodes are used.
- A weighted-distribution test that verifies keys are distributed proportionally to `Weight`.
- A concurrency test (run with `-race`) that hammers `GetNode`/`AddNode`/`RemoveNode` from many goroutines simultaneously.

## License

GPL-3.0