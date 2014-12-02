package route

import (
	"github.com/volkerp/goquadtree/quadtree"
)

type qtNode struct {
	c Coordinate
	i int
}

func (n *qtNode) BoundingBox() quadtree.BoundingBox {
	return qtBbox(n.c, 0)
}

func qtBbox(center Coordinate, radius float64) quadtree.BoundingBox {
	minX, maxX := center[0]-radius, center[0]+radius
	minY, maxY := center[1]-radius, center[1]+radius
	return quadtree.NewBoundingBox(minX, maxX, minY, maxY)
}

type qtNodeSorter struct {
	nodes  []quadtree.BoundingBoxer
	origin Coordinate
	coster Coster
}

func (s *qtNodeSorter) Len() int {
	return len(s.nodes)
}

func (s *qtNodeSorter) Less(i, j int) bool {
	cost := s.coster
	return cost(s.origin, s.nodes[i].(*qtNode).c) < cost(s.origin, s.nodes[j].(*qtNode).c)
}

func (s *qtNodeSorter) Swap(i, j int) {
	s.nodes[i], s.nodes[j] = s.nodes[j], s.nodes[i]
}
