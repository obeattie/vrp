package dag

import (
	"errors"
)

var (
	ErrCycle       = errors.New("Graph contains a cycle")
	ErrNodeMissing = errors.New("Node not found in graph")
)
