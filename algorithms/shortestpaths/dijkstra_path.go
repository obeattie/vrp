package shortestpaths

import (
	"errors"
	"math"

	"github.com/oleiade/lane"

	"github.com/obeattie/vrp/graph"
)

const priorityExponent = 100000.0

var (
	ErrUnreachable   = errors.New("Unreachable node")
	ErrContradiction = errors.New("Contradictory graph. Negative-cost edges?")
)

// DijkstraPath returns the shortest path from source to target.
// @TODO: Implement me
func DijkstraPath(g graph.Graph, source, target graph.Node) ([]graph.Node, error) {
	paths, _, err := singleSourceDijkstra(g, source, target, math.Inf(0))
	if err != nil {
		return nil, err
	} else if path, ok := paths[target]; ok {
		return path, nil
	} else {
		return nil, ErrUnreachable
	}
}

func singleSourceDijkstra(g graph.Graph, source, target graph.Node, cutoff float64) (map[graph.Node][]graph.Node, map[graph.Node]float64, error) {
	if source == target {
		paths := map[graph.Node][]graph.Node{
			source: {source},
		}
		costs := map[graph.Node]float64{
			source: 0,
		}
		return paths, costs, nil
	}

	costs := map[graph.Node]float64{}     // Dictionary of final costs
	paths := map[graph.Node][]graph.Node{ // Dictionary of paths
		source: {source},
	}
	seen := map[graph.Node]float64{
		source: 0.0,
	}
	fringe := lane.NewPQueue(lane.MINPQ)
	fringe.Push(source, 0)

	for fringe.Size() > 0 {
		_v, d := fringe.Pop()
		v := _v.(graph.Node)
		if _, ok := costs[v]; ok { // Already searched this node
			continue
		}
		costs[v] = float64(d) / priorityExponent
		if v == target {
			break
		}

		for _, w := range g.Successors(v) {
			edge := g.EdgeTo(v, w)
			vwDist := costs[v] + edge.Cost
			if vwDist > cutoff {
				continue
			}
			if wDist, ok := costs[w]; ok {
				if vwDist < wDist {
					return nil, nil, ErrContradiction
				}
			} else if wSeen, ok := seen[w]; !ok || vwDist < wSeen {
				seen[w] = vwDist
				fringe.Push(w, int(vwDist*priorityExponent))
				paths[w] = append(paths[v], w)
			}
		}
	}

	return paths, costs, nil
}
