package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/obeattie/vrp/graph"
)

type nodePrototype struct {
	srcId, targetId uint64
}

func TestTopologicalSort(t *testing.T) {
	suite.Run(t, new(TopologicalSortTestSuite))
}

type TopologicalSortTestSuite struct {
	suite.Suite
}

func (suite *TopologicalSortTestSuite) generateGraph(nodes []nodePrototype) graph.Graph {
	g := graph.NewGraph()

	for _, n := range nodes {
		g.AddDirectedEdge(&graph.Edge{
			H: &graph.Node{NodeId: n.srcId},
			T: &graph.Node{NodeId: n.targetId},
		})
	}

	return g
}

func (suite *TopologicalSortTestSuite) TestSort() {
	t := suite.T()
	g := suite.generateGraph([]nodePrototype{
		{1, 2},
		{2, 3},
		{3, 4},
		{4, 5},
		{4, 6},
		{5, 7},
		{6, 7},
	})

	nodes, err := TopologicalSort(g, false)
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes, 7)
}

func (suite *TopologicalSortTestSuite) TestSortCycles() {
	t := suite.T()
	g := suite.generateGraph([]nodePrototype{
		{1, 2},
		{2, 3},
		{3, 4},
		{4, 2},
	})

	nodes, err := TopologicalSort(g, false)
	assert.Error(t, err)
	assert.Nil(t, nodes)
}
