package consistenthashing

import (
	"crypto/md5"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var DefaultHashFunction HashFunc = func(key string) uint64 {
	sum := md5.Sum([]byte(key))
	return uint64(sum[0])<<56 | uint64(sum[1])<<48 | uint64(sum[2])<<40 | uint64(sum[3])<<32 | uint64(sum[4])<<24 | uint64(sum[5])<<16 | uint64(sum[6])<<8 | uint64(sum[7])
}

type consistentHash struct {
	mux              sync.RWMutex
	ring             []Node
	hashRing         map[uint64]struct{}
	hashFn           HashFunc
	baseVirtualNodes int
	virtualNodeCount map[string]int
}

func (c *consistentHash) GetNode(keyStr string) (Node, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if len(c.hashRing) == 0 {
		return Node{}, errors.New("consistent hash node not found")
	}

	keyHash := c.hashFn(keyStr)
	idx := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i].Hash >= keyHash
	})

	if idx == len(c.ring) {
		idx = 0
	}

	return c.ring[idx], nil
}

func (c *consistentHash) AddNode(node Node) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	key := c.hashFn(node.ID)
	if _, isExist := c.hashRing[key]; isExist {
		return errors.New("node already exists")
	}

	c.hashRing[key] = struct{}{}

	numVirtualNode := c.baseVirtualNodes * node.Weight
	newNodes := make([]Node, 0, numVirtualNode+1)
	realNode := node
	realNode.Hash = key
	newNodes = append(newNodes, realNode)

	var sb strings.Builder
	for i := 0; i < numVirtualNode; i++ {
		sb.Reset()
		sb.WriteString(node.ID)
		sb.WriteByte('#')
		sb.WriteString(strconv.Itoa(i))
		virtualNodeID := sb.String()

		virtualNodeHash := c.hashFn(virtualNodeID)

		vnode := node
		vnode.Hash = virtualNodeHash
		newNodes = append(newNodes, vnode)
	}
	c.virtualNodeCount[node.ID] = numVirtualNode

	sort.Slice(newNodes, func(i, j int) bool { return newNodes[i].Hash < newNodes[j].Hash })

	c.ring = mergeSortedRing(c.ring, newNodes)

	return nil
}

func (c *consistentHash) RemoveNode(node Node) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	key := c.hashFn(node.ID)

	if _, isExist := c.hashRing[key]; !isExist {
		return errors.New("node does not exist")
	}

	numVirtualNode := c.virtualNodeCount[node.ID]
	delete(c.virtualNodeCount, node.ID)
	delete(c.hashRing, key)

	hashesToRemove := map[uint64]struct{}{key: {}}
	var sb strings.Builder
	for i := 0; i < numVirtualNode; i++ {
		sb.Reset()
		sb.WriteString(node.ID)
		sb.WriteByte('#')
		sb.WriteString(strconv.Itoa(i))
		virtualNodeHash := c.hashFn(sb.String())
		hashesToRemove[virtualNodeHash] = struct{}{}
	}

	newRing := make([]Node, 0, len(c.ring))
	for _, n := range c.ring {
		if _, exists := hashesToRemove[n.Hash]; !exists {
			newRing = append(newRing, n)
		}
	}
	c.ring = newRing

	return nil
}

func NewConsistentHash(hashFunc HashFunc, baseVirtualNodes int) ConsistentHash {
	return &consistentHash{
		hashRing:         make(map[uint64]struct{}),
		ring:             make([]Node, 0),
		hashFn:           hashFunc,
		baseVirtualNodes: baseVirtualNodes,
		virtualNodeCount: make(map[string]int),
	}
}
