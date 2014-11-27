package shortestpaths

import (
	"errors"

	"github.com/oleiade/lane"

	"github.com/obeattie/vrp/graph"
)

const priorityExponent = 100000.0

// DijkstraPath returns the shortest path from source to target.
// @TODO: Implement me
func DijkstraPath(g graph.Graph, source, target *graph.Node) ([]*graph.Node, error) {
	return nil, nil
}

func singleSourceDijkstra(g graph.Graph, source, target *graph.Node, cutoff float64) (map[*graph.Node][]*graph.Node, map[*graph.Node]float64, error) {
	if source == target {
		paths := map[*graph.Node][]*graph.Node{
			source: {source},
		}
		distances := map[*graph.Node]float64{
			source: 0,
		}
		return paths, distances, nil
	}

	dist := map[*graph.Node]float64{}       // Dictionary of final distances
	paths := map[*graph.Node][]*graph.Node{ // Dictionary of paths
		source: {source},
	}
	seen := map[*graph.Node]float64{
		source: 0.0,
	}
	fringe := lane.NewPQueue(lane.MINPQ)
	fringe.Push(source, 0)

	for fringe.Size() > 0 {
		_v, d := fringe.Pop()
		v := _v.(*graph.Node)
		if _, ok := dist[v]; ok { // Already searched this node
			continue
		}
		dist[v] = float64(d) / priorityExponent
		if v == target {
			break
		}

		for _, w := range g.Successors(v) {
			edge := g.EdgeTo(v, w)
			vwDist := dist[v] + edge.Cost
			if vwDist > cutoff {
				continue
			}
			if wDist, ok := dist[w]; ok {
				if vwDist < wDist {
					return nil, nil, errors.New("Contradictory paths found. Negative costs?")
				}
			} else if wSeen, ok := seen[w]; !ok || vwDist < wSeen {
				seen[w] = vwDist
				fringe.Push(w, int(vwDist*priorityExponent))
				paths[w] = append(paths[v], w)
			}
		}
	}

	return paths, dist, nil
}
