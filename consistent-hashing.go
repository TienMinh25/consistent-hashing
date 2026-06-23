package consistenthashing

import (
	"errors"
	"hash/crc32"
	"sort"
)

var DefaultHashFunction HashFunc = func(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

type Node struct {
	ID string
}

type ConsistentHash interface {
	GetNode(keyStr string) (Node, error)
	AddNode(node Node) error
	RemoveNode(node Node) error
}

type HashFunc func(string) uint32

type consistentHash struct {
	hashToNode   map[uint32]Node
	sortedHashes []uint32
	hashFn       HashFunc
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

	return c.hashToNode[nodeHashFind], nil
}

func (c *consistentHash) AddNode(node Node) error {
	key := c.hashFn(node.ID)
	if _, isExist := c.hashToNode[key]; isExist {
		return errors.New("node already exists")
	}

	c.hashToNode[key] = node
	c.sortedHashes = append(c.sortedHashes, key)
	sort.Slice(c.sortedHashes, func(i, j int) bool { return c.sortedHashes[i] < c.sortedHashes[j] })
	return nil
}

func (c *consistentHash) RemoveNode(node Node) error {
	key := c.hashFn(node.ID)

	if _, isExist := c.hashToNode[key]; isExist {
		delete(c.hashToNode, key)

		sortedHashes := make([]uint32, 0, len(c.sortedHashes)-1)
		for _, hash := range c.sortedHashes {
			if hash != key {
				sortedHashes = append(sortedHashes, hash)
			}
		}
		c.sortedHashes = sortedHashes
	}

	return nil
}

func NewConsistentHash(hashFunc HashFunc) ConsistentHash {
	return &consistentHash{
		hashToNode:   make(map[uint32]Node),
		sortedHashes: make([]uint32, 0),
		hashFn:       hashFunc,
	}
}
