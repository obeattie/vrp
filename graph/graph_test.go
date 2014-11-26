package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGraph(t *testing.T) {
	suite.Run(t, new(GraphTestSuite))
}

type GraphTestSuite struct {
	suite.Suite
	g RouteGraph
}

func (suite *GraphTestSuite) SetupTest() {
	g := NewGraph()
	nodes := [...]struct {
		srcId, targetId uint64
		cost            float64
	}{
		{2, 2, 2},
		{1, 0, 3},
		{2, 0, 1},
		{0, 2, 2},
	}

	for _, n := range nodes {
		g.AddDirectedEdge(&Edge{
			H: &Node{
				NodeId: n.srcId,
			},
			T: &Node{
				NodeId: n.targetId,
			},
			Cost: n.cost,
		})
	}

	suite.g = g
}

func (suite *GraphTestSuite) TestNodeExists() {
	t, g := suite.T(), suite.g
	for _, n := range g.NodeList() {
		assert.True(t, g.NodeExists(n))
	}
}

func (suite *GraphTestSuite) TestNodeList() {
	t, g := suite.T(), suite.g
	assert.Len(t, g.NodeList(), 3) // There are three unique nodes in the graph, by id
}

func (suite *GraphTestSuite) TestNeighbors() {
	t, g := suite.T(), suite.g
	for _, n := range g.NodeList() {
		switch n.NodeId {
		case 0:
			assert.Len(t, g.Neighbors(n), 2)
		case 1:
			assert.Len(t, g.Neighbors(n), 1)
		case 2:
			assert.Len(t, g.Neighbors(n), 2) // Node 2 is its own neighour
		default:
			assert.Fail(t, "Do not want")
		}
	}
}

func (suite *GraphTestSuite) EdgeBetween() {
	t, g := suite.T(), suite.g

	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 2}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 0}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 0}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 2}))

	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 1}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 1}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 2}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 0}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 1}))
}
