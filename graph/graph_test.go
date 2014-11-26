package graph

import (
	"fmt"
	"sync"
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
			H:    &Node{NodeId: n.srcId},
			T:    &Node{NodeId: n.targetId},
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

func (suite *GraphTestSuite) TestEdgeBetween() {
	t, g := suite.T(), suite.g

	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 0}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 1}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 2}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 0}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 1}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 2}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 0}))
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 1}))
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 2}))
}

func (suite *GraphTestSuite) TestEdgeBetweenWeights() {
	t, g := suite.T(), suite.g

	assert.Equal(t, 3, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 0}).Cost)
	assert.Equal(t, 3, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 1}).Cost)
	assert.Equal(t, 2, g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 2}).Cost)
	assert.Equal(t, 1, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 0}).Cost)
	assert.Equal(t, 2, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 2}).Cost)
}

func (suite *GraphTestSuite) TestSuccessors() {
	t, g := suite.T(), suite.g

	assert.Len(t, g.Successors(&Node{NodeId: 0}), 1)
	assert.Len(t, g.Successors(&Node{NodeId: 1}), 1)
	assert.Len(t, g.Successors(&Node{NodeId: 2}), 2)
}

func (suite *GraphTestSuite) TestEdgeTo() {
	t, g := suite.T(), suite.g

	assert.Nil(t, g.EdgeTo(&Node{NodeId: 0}, &Node{NodeId: 0}))
	assert.Nil(t, g.EdgeTo(&Node{NodeId: 0}, &Node{NodeId: 1}))
	assert.NotNil(t, g.EdgeTo(&Node{NodeId: 0}, &Node{NodeId: 2}))
	assert.NotNil(t, g.EdgeTo(&Node{NodeId: 1}, &Node{NodeId: 0}))
	assert.Nil(t, g.EdgeTo(&Node{NodeId: 1}, &Node{NodeId: 1}))
	assert.Nil(t, g.EdgeTo(&Node{NodeId: 1}, &Node{NodeId: 2}))
	assert.NotNil(t, g.EdgeTo(&Node{NodeId: 2}, &Node{NodeId: 0}))
	assert.Nil(t, g.EdgeTo(&Node{NodeId: 2}, &Node{NodeId: 1}))
	assert.NotNil(t, g.EdgeTo(&Node{NodeId: 2}, &Node{NodeId: 2}))
}

func (suite *GraphTestSuite) TestEdgeToWeights() {
	t, g := suite.T(), suite.g

	assert.Equal(t, 2, g.EdgeTo(&Node{NodeId: 0}, &Node{NodeId: 2}).Cost)
	assert.Equal(t, 3, g.EdgeTo(&Node{NodeId: 1}, &Node{NodeId: 0}).Cost)
	assert.Equal(t, 1, g.EdgeTo(&Node{NodeId: 2}, &Node{NodeId: 0}).Cost)
	assert.Equal(t, 2, g.EdgeTo(&Node{NodeId: 2}, &Node{NodeId: 2}).Cost)
}

func (suite *GraphTestSuite) TestPredecessors() {
	t, g := suite.T(), suite.g

	assert.Len(t, g.Predecessors(&Node{NodeId: 0}), 2)
	assert.Len(t, g.Predecessors(&Node{NodeId: 1}), 0)
	assert.Len(t, g.Predecessors(&Node{NodeId: 2}), 2)
}

func (suite *GraphTestSuite) TestCost() {
	t, g := suite.T(), suite.g

	e := g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 0})
	assert.Equal(t, g.Cost(e), e.Cost)
	e = g.EdgeBetween(&Node{NodeId: 0}, &Node{NodeId: 2})
	assert.Equal(t, g.Cost(e), e.Cost)
	e = g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 2})
	assert.Equal(t, g.Cost(e), e.Cost)
	e = g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 0})
	assert.Equal(t, g.Cost(e), e.Cost)
}

func (suite *GraphTestSuite) TestCostNil() {
	t, g := suite.T(), suite.g

	assert.NotPanics(t, func() {
		g.Cost(nil)
	})
}

func (suite *GraphTestSuite) TestNewNode() {
	t, g := suite.T(), suite.g
	assert.NotNil(t, g.NewNode())
}

func (suite *GraphTestSuite) TestNewNodeNoClashes() {
	t, g := suite.T(), suite.g
	if testing.Short() {
		t.Skipf("Skipped in short mode")
	}

	workers := 200
	nodesPerWorker := 2500

	barrier := new(sync.WaitGroup)
	barrier.Add(workers)
	done := new(sync.WaitGroup)
	done.Add(workers)
	workerResults := make([]map[int]bool, workers)

	worker := func(workerId int) {
		defer done.Done()
		resultIds := make(map[int]bool, nodesPerWorker)
		barrier.Done()
		barrier.Wait()

		for i := 0; i < nodesPerWorker; i++ {
			n := g.NewNode()

			if _, ok := resultIds[n.ID()]; ok {
				assert.Fail(t, "Conflicting ID %s", n.ID())
				break
			} else {
				resultIds[n.ID()] = true
			}
		}

		workerResults[workerId] = resultIds
	}

	for i := 0; i < workers; i++ {
		go worker(i)
	}
	done.Wait()

	allIds := make(map[int]bool, nodesPerWorker*workers)
	for _, idSet := range workerResults {
		for nodeId, _ := range idSet {
			if _, ok := allIds[nodeId]; ok {
				assert.Fail(t, fmt.Sprintf("Conflicting ID %d", nodeId))
				break
			} else {
				allIds[nodeId] = true
			}
		}
	}
}

func (suite *GraphTestSuite) TestAddNode() {
	t, g := suite.T(), suite.g

	g.AddNode(&Node{NodeId: 4})
	assert.Len(t, g.NodeList(), 4)
}

func (suite *GraphTestSuite) TestAddDuplicateNode() {
	t, g := suite.T(), suite.g

	g.AddNode(&Node{NodeId: 4})
	assert.Len(t, g.NodeList(), 4)
	g.AddNode(&Node{NodeId: 4}) // Doing it again shouldn't increase the node count
	assert.Len(t, g.NodeList(), 4)
}

func (suite *GraphTestSuite) TestRemoveNode() {
	t, g := suite.T(), suite.g

	g.RemoveNode(&Node{NodeId: 2})
	assert.Len(t, g.NodeList(), 2)
}

func (suite *GraphTestSuite) TestNonexistentRemoveNode() {
	t, g := suite.T(), suite.g

	g.RemoveNode(&Node{NodeId: 4})
	assert.Len(t, g.NodeList(), 3)
}

func (suite *GraphTestSuite) TestAddDirectedEdge() {
	t, g := suite.T(), suite.g

	g.AddDirectedEdge(&Edge{
		T:    &Node{NodeId: 1},
		H:    &Node{NodeId: 1},
		Cost: 1,
	})
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 1}))
}

func (suite *GraphTestSuite) TestAddDuplicateDirectedEdge() {
	t, g := suite.T(), suite.g

	g.AddDirectedEdge(&Edge{
		T:    &Node{NodeId: 2},
		H:    &Node{NodeId: 2},
		Cost: 1,
	})
	assert.NotNil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 2}))
}

func (suite *GraphTestSuite) TestRemoveDirectedEdge() {
	t, g := suite.T(), suite.g

	g.RemoveDirectedEdge(&Edge{
		T:    &Node{NodeId: 2},
		H:    &Node{NodeId: 2},
		Cost: 1,
	})
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 2}, &Node{NodeId: 2}))
}

func (suite *GraphTestSuite) TestRemoveNonexistentDirectedEdge() {
	t, g := suite.T(), suite.g

	g.RemoveDirectedEdge(&Edge{
		T:    &Node{NodeId: 1},
		H:    &Node{NodeId: 1},
		Cost: 1,
	})
	assert.Nil(t, g.EdgeBetween(&Node{NodeId: 1}, &Node{NodeId: 1}))
}
