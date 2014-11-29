package graph

import (
	"math"
	"sync"
	"sync/atomic"

	graphlib "github.com/gonum/graph"
	concretegraphlib "github.com/gonum/graph/concrete"
)

type Graph interface {
	// Essentially implements versions of the following interfaces that return our own graph primitives:
	// - graphlib.MutableDirectedGraph
	//   - graphlib.CostDirectedGraph
	//     - graphlib.Coster
	//     - graphlib.DirectedGraph
	//       - graphlib.Graph
	//   - graphlib.Mutable

	// graphlib.Graph

	NodeExists(Node) bool
	NodeList() []Node
	Neighbors(Node) []Node
	EdgeBetween(node, neighbour Node) *Edge

	// graphlib.DirectedGraph

	Successors(Node) []Node
	EdgeTo(node, successor Node) *Edge
	Predecessors(Node) []Node

	// graphlib.Coster

	Cost(*Edge) float64

	// graphlib.Mutable

	NewNode() Node
	AddNode(Node)
	RemoveNode(Node)

	// graphlib.MutableDirectedGraph

	AddDirectedEdge(e *Edge)
	RemoveDirectedEdge(e *Edge)

	Copy() Graph
}

type graphImpl struct {
	sync.RWMutex
	g         graphlib.MutableDirectedGraph
	nodeIdSeq *uint64 // Atomically updated
}

func NewGraph() Graph {
	_one := uint64(1)
	return &graphImpl{
		g:         concretegraphlib.NewDirectedGraph(),
		nodeIdSeq: &_one,
	}
}

func (g *graphImpl) graphNodeToNode(n graphlib.Node) Node {
	if result, ok := n.(Node); ok {
		return result
	}
	return Node{}
}

func (g *graphImpl) graphNodesToNodes(n []graphlib.Node) []Node {
	if n == nil {
		return nil
	}

	result := make([]Node, 0, len(n))
	for _, candidate := range n {
		if resultNode := g.graphNodeToNode(candidate); !resultNode.IsZero() {
			result = append(result, resultNode)
		}
	}
	return result
}

func (g *graphImpl) graphEdgeToEdge(e graphlib.Edge) *Edge {
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

func (g *graphImpl) graphEdgesToEdges(e []graphlib.Edge) []*Edge {
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

func (g *graphImpl) NodeExists(n Node) bool {
	g.RLock()
	defer g.RUnlock()

	return g.g.NodeExists(n)
}

func (g *graphImpl) NodeList() []Node {
	g.RLock()
	defer g.RUnlock()

	return g.graphNodesToNodes(g.g.NodeList())
}

func (g *graphImpl) Neighbors(n Node) []Node {
	g.RLock()
	defer g.RUnlock()

	return g.graphNodesToNodes(g.g.Neighbors(n))
}

func (g *graphImpl) EdgeBetween(n, neigh Node) *Edge {
	g.RLock()
	defer g.RUnlock()

	return g.graphEdgeToEdge(g.g.EdgeBetween(n, neigh))
}

func (g *graphImpl) Successors(n Node) []Node {
	g.RLock()
	defer g.RUnlock()

	return g.graphNodesToNodes(g.g.Successors(n))
}

func (g *graphImpl) EdgeTo(node, successor Node) *Edge {
	g.RLock()
	defer g.RUnlock()

	return g.graphEdgeToEdge(g.g.EdgeTo(node, successor))
}

func (g *graphImpl) Predecessors(n Node) []Node {
	g.RLock()
	defer g.RUnlock()

	return g.graphNodesToNodes(g.g.Predecessors(n))
}

func (g *graphImpl) Cost(e *Edge) float64 {
	g.RLock()
	defer g.RUnlock()

	if e == nil {
		return math.Inf(0)
	}
	return e.Cost
}

func (g *graphImpl) generateNodeId() int {
	return int(atomic.AddUint64(g.nodeIdSeq, 1))
}

func (g *graphImpl) NewNode() Node {
	n := Node{Id: g.generateNodeId()}
	g.AddNode(n)
	return n
}

func (g *graphImpl) AddNode(n Node) {
	g.Lock()
	defer g.Unlock()

	g.g.AddNode(n)
}

func (g *graphImpl) RemoveNode(n Node) {
	g.Lock()
	defer g.Unlock()

	g.g.RemoveNode(n)
}

func (g *graphImpl) AddDirectedEdge(e *Edge) {
	g.Lock()
	defer g.Unlock()

	g.g.AddDirectedEdge(e, e.Cost)
}

func (g *graphImpl) RemoveDirectedEdge(e *Edge) {
	g.Lock()
	defer g.Unlock()

	g.g.RemoveDirectedEdge(e)
}

func (g *graphImpl) Copy() Graph {
	result := NewGraph()
	nodes := g.NodeList()
	for _, n := range nodes {
		result.AddNode(n)
		for _, predecessor := range g.Predecessors(n) {
			result.AddDirectedEdge(g.EdgeTo(predecessor, n))
		}
	}
	return result
}
