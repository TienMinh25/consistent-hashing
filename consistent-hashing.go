package consistenthashing

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
)

var DefaultHashFunction HashFunc = func(key string) uint32 {
	sum := md5.Sum([]byte(key))
	return uint32(sum[0])<<24 | uint32(sum[1])<<16 | uint32(sum[2])<<8 | uint32(sum[3])
}

type Node struct {
	ID     string
	Weight int
}

type ConsistentHash interface {
	GetNode(keyStr string) (Node, error)
	AddNode(node Node) error
	RemoveNode(node Node) error
}

type HashFunc func(string) uint32

type consistentHash struct {
	hashToNode        map[uint32]Node
	sortedHashes      []uint32
	hashFn            HashFunc
	virtualNodeToNode map[uint32]Node
	baseVirtualNodes  int
	virtualNodeCount  map[string]int
}

func (c *consistentHash) GetNode(keyStr string) (Node, error) {
	if len(c.hashToNode) == 0 {
		return Node{}, errors.New("consistent hash node not found")
	}

	keyHash := c.hashFn(keyStr)
	idx := sort.Search(len(c.sortedHashes), func(i int) bool {
		return c.sortedHashes[i] >= keyHash
	})

	var nodeHashFind uint32
	if idx == len(c.sortedHashes) {
		nodeHashFind = c.sortedHashes[0]
	} else {
		nodeHashFind = c.sortedHashes[idx]
	}

	if node, isExist := c.virtualNodeToNode[nodeHashFind]; isExist {
		return node, nil
	}

	return c.hashToNode[nodeHashFind], nil
}

func (c *consistentHash) AddNode(node Node) error {
	key := c.hashFn(node.ID)
	if _, isExist := c.hashToNode[key]; isExist {
		return errors.New("node already exists")
	}

	c.hashToNode[key] = node
	c.sortedHashes = append(c.sortedHashes, key)
	
	numVirtualNode := c.baseVirtualNodes * node.Weight
	for i := 0; i < numVirtualNode; i++ {
		virtualNodeID := fmt.Sprintf("%s#%d", node.ID, i)
		virtualNodeHash := c.hashFn(virtualNodeID)
		c.virtualNodeToNode[virtualNodeHash] = node
		c.sortedHashes = append(c.sortedHashes, virtualNodeHash)
	}
	c.virtualNodeCount[node.ID] = numVirtualNode

	sort.Slice(c.sortedHashes, func(i, j int) bool { return c.sortedHashes[i] < c.sortedHashes[j] })
	return nil
}

func (c *consistentHash) RemoveNode(node Node) error {
	key := c.hashFn(node.ID)

	if _, isExist := c.hashToNode[key]; isExist {
		numVirtualNode := c.virtualNodeCount[node.ID]
		delete(c.virtualNodeCount, node.ID)
		delete(c.hashToNode, key)

		hashesToRemove := map[uint32]struct{}{
			key: {},
		}
		for i := 0; i < numVirtualNode; i++ {
			virtualNodeID := fmt.Sprintf("%s#%d", node.ID, i)
			virtualNodeHash := c.hashFn(virtualNodeID)
			hashesToRemove[virtualNodeHash] = struct{}{}
			delete(c.virtualNodeToNode, virtualNodeHash)
		}

		sortedHashes := make([]uint32, 0, len(c.sortedHashes)-1)
		for _, hash := range c.sortedHashes {
			if _, exists := hashesToRemove[hash]; !exists {
				sortedHashes = append(sortedHashes, hash)
			}
		}
		c.sortedHashes = sortedHashes
	}

	return nil
}

func NewConsistentHash(hashFunc HashFunc, baseVirtualNodes int) ConsistentHash {
	return &consistentHash{
		hashToNode:        make(map[uint32]Node),
		sortedHashes:      make([]uint32, 0),
		hashFn:            hashFunc,
		virtualNodeToNode: make(map[uint32]Node),
		baseVirtualNodes:  baseVirtualNodes,
		virtualNodeCount:  make(map[string]int),
	}
}
