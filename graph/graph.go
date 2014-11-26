package graph

import (
	"sync/atomic"

	graphlib "github.com/gonum/graph"
	concretegraphlib "github.com/gonum/graph/concrete"
)

type RouteGraph interface {
	// Essentially implement versions of the following interfaces that return our own graph primitives:
	// - graphlib.MutableDirectedGraph
	//   - graphlib.CostDirectedGraph
	//     - graphlib.Coster
	//     - graphlib.DirectedGraph
	//       - graphlib.Graph
	//   - graphlib.Mutable

	// graphlib.Graph
	NodeExists(*Node) bool
	NodeList() []*Node
	Neighbors(*Node) []*Node
	EdgeBetween(node, neighbour *Node) *Edge

	// graphlib.DirectedGraph
	Successors(*Node) []*Node
	EdgeTo(node, successor *Node) *Edge
	Predecessors(*Node) []*Node

	// graphlib.Coster
	Cost(*Edge) float64

	// graphlib.Mutable
	NewNode() *Node
	AddNode(*Node)
	RemoveNode(*Node)

	// graphlib.MutableDirectedGraph
	AddDirectedEdge(e *Edge)
	RemoveDirectedEdge(e *Edge)
}

type routeGraphImpl struct {
	g         graphlib.MutableDirectedGraph
	nodeIdSeq *uint64 // Atomically updated
}

func NewGraph() RouteGraph {
	_zero := uint64(0)
	return &routeGraphImpl{
		g:         concretegraphlib.NewDirectedGraph(),
		nodeIdSeq: &_zero,
	}
}

// routeGraphImpl utilities

func (g *routeGraphImpl) graphNodeToNode(n graphlib.Node) *Node {
	if result, ok := n.(*Node); ok {
		return result
	}
	return nil
}

func (g *routeGraphImpl) graphNodesToNodes(n []graphlib.Node) []*Node {
	if n == nil {
		return nil
	}

	result := make([]*Node, 0, len(n))
	for _, candidate := range n {
		if resultNode := g.graphNodeToNode(candidate); resultNode != nil {
			result = append(result, resultNode)
		}
	}
	return result
}

func (g *routeGraphImpl) graphEdgeToEdge(e graphlib.Edge) *Edge {
	if result, ok := e.(*Edge); ok {
		return result
	} else if we, ok := (e.(concretegraphlib.WeightedEdge)); ok {
		return &Edge{
			H:    g.graphNodeToNode(we.Head()),
			T:    g.graphNodeToNode(we.Tail()),
			Cost: we.Cost,
		}
	}
	return nil
}

func (g *routeGraphImpl) graphEdgesToEdges(e []graphlib.Edge) []*Edge {
	if e == nil {
		return nil
	}

	result := make([]*Edge, 0, len(e))
	for _, candidate := range e {
		if resultEdge := g.graphEdgeToEdge(candidate); candidate != nil {
			result = append(result, resultEdge)
		}
	}
	return result
}

func (g *routeGraphImpl) NodeExists(n *Node) bool {
	return g.g.NodeExists(n)
}

func (g *routeGraphImpl) NodeList() []*Node {
	return g.graphNodesToNodes(g.g.NodeList())
}

func (g *routeGraphImpl) Neighbors(n *Node) []*Node {
	return g.graphNodesToNodes(g.g.Neighbors(n))
}

func (g *routeGraphImpl) EdgeBetween(n, neigh *Node) *Edge {
	return g.graphEdgeToEdge(g.g.EdgeBetween(n, neigh))
}

func (g *routeGraphImpl) Successors(n *Node) []*Node {
	return g.graphNodesToNodes(g.g.Successors(n))
}

func (g *routeGraphImpl) EdgeTo(node, successor *Node) *Edge {
	return g.graphEdgeToEdge(g.g.EdgeTo(node, successor))
}

func (g *routeGraphImpl) Predecessors(n *Node) []*Node {
	return g.graphNodesToNodes(g.g.Predecessors(n))
}

func (g *routeGraphImpl) Cost(e *Edge) float64 {
	if e == nil {
		return math.Inf(0)
	}
	return e.Cost
}

func (g *routeGraphImpl) generateNodeId() uint64 {
	return atomic.AddUint64(g.nodeIdSeq, 1)
}

func (g *routeGraphImpl) NewNode() *Node {
	n := &Node{NodeId: g.generateNodeId()}
	g.AddNode(n)
	return n
}

func (g *routeGraphImpl) AddNode(n *Node) {
	g.g.AddNode(n)
}

func (g *routeGraphImpl) RemoveNode(n *Node) {
	g.g.RemoveNode(n)
}

func (g *routeGraphImpl) AddDirectedEdge(e *Edge) {
	g.g.AddDirectedEdge(e, e.Cost)
}

func (g *routeGraphImpl) RemoveDirectedEdge(e *Edge) {
	g.g.RemoveDirectedEdge(e)
}
