package graph

import (
	graphlib "github.com/gonum/graph"
)

type Edge struct {
	H, T *Node
	Cost float64
}

func (e Edge) Head() graphlib.Node {
	return e.H
}

func (e Edge) Tail() graphlib.Node {
	return e.T
}
