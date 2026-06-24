package consistenthashing

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