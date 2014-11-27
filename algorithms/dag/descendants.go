package dag

import (
	"github.com/obeattie/vrp/graph"
)

func Descendants(g graph.Graph, origin graph.Node) ([]graph.Node, error) {
	if !g.NodeExists(origin) {
		return nil, ErrNodeMissing
	}

	var n graph.Node
	resultSet := make(map[int]graph.Node, 10)
	toVisit := []graph.Node{origin}
	for len(toVisit) > 0 {
		n, toVisit = toVisit[0], toVisit[1:]
		if _, ok := resultSet[n.ID()]; ok { // Already visited
			continue
		}
		resultSet[n.ID()] = n
		toVisit = append(toVisit, g.Successors(n)...)
	}

	// Build a slice of the resultSet
	results := make([]graph.Node, 0, len(toVisit))
	for _, n := range resultSet {
		if n.ID() != origin.ID() {
			results = append(results, n)
		}
	}
	return results, nil
}
