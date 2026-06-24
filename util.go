package consistenthashing

func mergeSortedRing(ring1, ring2 []Node) []Node {
	merged := make([]Node, 0, len(ring1)+len(ring2))
	i, j := 0, 0

	for i < len(ring1) && j < len(ring2) {
		if ring1[i].Hash < ring2[j].Hash {
			merged = append(merged, ring1[i])
			i++
		} else {
			merged = append(merged, ring2[j])
			j++
		}
	}

	for i < len(ring1) {
		merged = append(merged, ring1[i])
		i++
	}

	for j < len(ring2) {
		merged = append(merged, ring2[j])
		j++
	}

	return merged
}
