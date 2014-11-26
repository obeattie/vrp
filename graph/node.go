package graph

type Node struct {
	NodeId uint64
	Lat    float64
	Lng    float64
}

func (n *Node) ID() int {
	return int(n.NodeId)
}
