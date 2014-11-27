package graph

type Node struct {
	Id  int
	Lat float64
	Lng float64
}

func (n Node) ID() int {
	return n.Id
}

func (n Node) IsZero() bool {
	return n.Id == 0
}
