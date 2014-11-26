package dag

import (
	"errors"

	"github.com/obeattie/vrp/graph"
)

var (
	ErrCycle = errors.New("Graph contains a cycle")
)

// TopologicalSort returns a list of Nodes in topological sort order.
//
// A topological sort is a noninique permutation of the nodes such that an edge from u to v implies that u appears
// before v in the topological sort order.
//
// If a topological sort is infeasible because the given Graph contains cycles, ErrCycle is returned.
//
// This algorithm is based on a description and proof in The Algorithm Design Manual [1].
//
// [1] Skiena, S. S. The Algorithm Design Manual  (Springer-Verlag, 1998).
//     http://www.amazon.com/exec/obidos/ASIN/0387948600/ref=ase_thealgorithmrepo/
func TopologicalSort(g graph.Graph) ([]*graph.Node, error) {
	order, err := TopologicalSortReverse(g)
	if err != nil {
		return order, err
	}

	sorted := make([]*graph.Node, len(order))
	for i, newI := len(order)-1, 0; i >= 0; i, newI = i-1, newI+1 {
		sorted[newI] = order[i]
	}
	return sorted, nil
}

// TopologicalSortReverse returns a postorder topological sort of the Nodes (ie. an array in the reverse order to that
// returned by TopologicalSort).
func TopologicalSortReverse(g graph.Graph) ([]*graph.Node, error) {
	nodesList := g.NodeList()
	seen := make(map[*graph.Node]bool)
	order := make([]*graph.Node, 0, len(nodesList))
	explored := make(map[*graph.Node]bool)

	for _, v := range nodesList {
		if _, ok := explored[v]; ok { // Node has been explored already
			continue
		}

		fringe := []*graph.Node{v}
		for len(fringe) > 0 {
			w := fringe[len(fringe)-1]
			if _, ok := explored[w]; ok { // Node has been explored already
				fringe = fringe[:len(fringe)-1]
				continue
			}
			seen[w] = true // Mark as seen

			// Check successors for cycles and for new nodes
			new_nodes := make([]*graph.Node, 0)
			for _, n := range g.Successors(w) {
				if _, ok := explored[n]; !ok {
					if _, ok = seen[n]; ok { // Cycle!
						return nil, ErrCycle
					}
					new_nodes = append(new_nodes, n)
				}
			}
			if len(new_nodes) > 0 { // Add new_nodes to fringe
				fringe = append(fringe, new_nodes...)
			} else { // No new nodes, so fringe is fully explored
				explored[w] = true
				order = append(order, w)
				fringe = fringe[:len(fringe)-1] // We're done considering this node
			}
		}
	}

	return order, nil
}
