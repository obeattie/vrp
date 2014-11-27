package shortestpaths

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/obeattie/vrp/graph"
)

type nodePrototype struct {
	srcId, targetId uint64
}

func TestDijkstraPath(t *testing.T) {
	suite.Run(t, new(DijkstraPathTestSuite))
}

type DijkstraPathTestSuite struct {
	suite.Suite
}

func (suite *DijkstraPathTestSuite) generateGraph(nodes []nodePrototype, costs float64) graph.Graph {
	g := graph.NewGraph()

	for _, n := range nodes {
		g.AddDirectedEdge(&graph.Edge{
			H:    &graph.Node{NodeId: n.srcId},
			T:    &graph.Node{NodeId: n.targetId},
			Cost: costs,
		})
	}

	return g
}

func (suite *DijkstraPathTestSuite) TestSingleSourceDijkstra() {
	t := suite.T()
	g := suite.generateGraph([]nodePrototype{
		{1, 2},
		{2, 3},
		{3, 4},
		{4, 5},
		{4, 6},
		{5, 7},
		{6, 7},
	}, 2)

	paths, costs, err := singleSourceDijkstra(g, &graph.Node{NodeId: 1}, nil, math.Inf(0))
	assert.NoError(t, err)
	assert.Len(t, paths, 7) // Should include a path to itsel
	assert.Len(t, costs, 7)

	validPaths := map[int][][]int{
		1: {{1}},
		2: {{1, 2}},
		3: {{1, 2, 3}},
		4: {{1, 2, 3, 4}},
		5: {{1, 2, 3, 4, 5}},
		6: {{1, 2, 3, 4, 5, 6}},
		7: {{1, 2, 3, 4, 5, 7},
			{1, 2, 3, 4, 6, 7}},
	}

	matched := false

destinationLoop:
	for destination, path := range paths {
	candidatePathLoop:
		for _, candidatePath := range validPaths[destination.ID()] {
			for i, pathNode := range path {
				if i > len(candidatePath)-1 || pathNode.ID() != candidatePath[i] {
					continue candidatePathLoop
				}
			}
			matched = true
			break destinationLoop
		}
	}

	assert.True(t, matched)

	for destination, cost := range costs {
		assert.Equal(t, (len(paths[destination])-1)*2, cost)
	}
}
