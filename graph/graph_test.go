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
	g Graph
}

func (suite *GraphTestSuite) SetupTest() {
	g := NewGraph()
	nodes := [...]struct {
		srcId, targetId int
		cost            float64
	}{
		{3, 3, 2},
		{2, 1, 3},
		{3, 1, 1},
		{1, 3, 2},
	}

	for _, n := range nodes {
		g.AddDirectedEdge(&Edge{
			H:    Node{Id: n.srcId},
			T:    Node{Id: n.targetId},
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
	neighbours := map[int][]int{
		1: {2, 3},
		2: {1},
		3: {1, 3},
	}

	for _, n := range g.NodeList() {
		assert.Len(t, g.Neighbors(n), len(neighbours[n.ID()]), "Invalid neighbours for node %d", n.ID())
	}
}

func (suite *GraphTestSuite) TestEdgeBetween() {
	t, g := suite.T(), suite.g

	assert.Nil(t, g.EdgeBetween(Node{Id: 1}, Node{Id: 1}))
	assert.NotNil(t, g.EdgeBetween(Node{Id: 1}, Node{Id: 2}))
	assert.NotNil(t, g.EdgeBetween(Node{Id: 1}, Node{Id: 3}))
	assert.NotNil(t, g.EdgeBetween(Node{Id: 2}, Node{Id: 1}))
	assert.Nil(t, g.EdgeBetween(Node{Id: 2}, Node{Id: 2}))
	assert.Nil(t, g.EdgeBetween(Node{Id: 2}, Node{Id: 3}))
	assert.NotNil(t, g.EdgeBetween(Node{Id: 3}, Node{Id: 1}))
	assert.Nil(t, g.EdgeBetween(Node{Id: 3}, Node{Id: 2}))
	assert.NotNil(t, g.EdgeBetween(Node{Id: 3}, Node{Id: 3}))
}

func (suite *GraphTestSuite) TestEdgeBetweenWeights() {
	t, g := suite.T(), suite.g

	assert.Equal(t, 3, g.EdgeBetween(Node{Id: 2}, Node{Id: 1}).Cost)
	assert.Equal(t, 3, g.EdgeBetween(Node{Id: 1}, Node{Id: 2}).Cost)
	assert.Equal(t, 2, g.EdgeBetween(Node{Id: 1}, Node{Id: 3}).Cost)
	assert.Equal(t, 1, g.EdgeBetween(Node{Id: 3}, Node{Id: 1}).Cost)
	assert.Equal(t, 2, g.EdgeBetween(Node{Id: 3}, Node{Id: 3}).Cost)
}

func (suite *GraphTestSuite) TestSuccessors() {
	t, g := suite.T(), suite.g

	assert.Len(t, g.Successors(Node{Id: 1}), 1)
	assert.Len(t, g.Successors(Node{Id: 2}), 1)
	assert.Len(t, g.Successors(Node{Id: 3}), 2)
}

func (suite *GraphTestSuite) TestEdgeTo() {
	t, g := suite.T(), suite.g

	assert.Nil(t, g.EdgeTo(Node{Id: 1}, Node{Id: 1}))
	assert.Nil(t, g.EdgeTo(Node{Id: 1}, Node{Id: 2}))
	assert.NotNil(t, g.EdgeTo(Node{Id: 1}, Node{Id: 3}))
	assert.NotNil(t, g.EdgeTo(Node{Id: 2}, Node{Id: 1}))
	assert.Nil(t, g.EdgeTo(Node{Id: 2}, Node{Id: 3}))
	assert.Nil(t, g.EdgeTo(Node{Id: 2}, Node{Id: 3}))
	assert.NotNil(t, g.EdgeTo(Node{Id: 2}, Node{Id: 1}))
	assert.Nil(t, g.EdgeTo(Node{Id: 3}, Node{Id: 2}))
	assert.NotNil(t, g.EdgeTo(Node{Id: 3}, Node{Id: 3}))
}

func (suite *GraphTestSuite) TestEdgeToWeights() {
	t, g := suite.T(), suite.g

	assert.Equal(t, 2, g.EdgeTo(Node{Id: 1}, Node{Id: 3}).Cost)
	assert.Equal(t, 3, g.EdgeTo(Node{Id: 2}, Node{Id: 1}).Cost)
	assert.Equal(t, 1, g.EdgeTo(Node{Id: 3}, Node{Id: 1}).Cost)
	assert.Equal(t, 2, g.EdgeTo(Node{Id: 3}, Node{Id: 3}).Cost)
}

func (suite *GraphTestSuite) TestPredecessors() {
	t, g := suite.T(), suite.g

	assert.Len(t, g.Predecessors(Node{Id: 1}), 2)
	assert.Len(t, g.Predecessors(Node{Id: 2}), 0)
	assert.Len(t, g.Predecessors(Node{Id: 3}), 2)
}

func (suite *GraphTestSuite) TestCost() {
	t, g := suite.T(), suite.g

	e := g.EdgeBetween(Node{Id: 2}, Node{Id: 1})
	assert.NotNil(t, e)
	assert.Equal(t, g.Cost(e), e.Cost)
	e = g.EdgeBetween(Node{Id: 1}, Node{Id: 3})
	assert.NotNil(t, e)
	assert.Equal(t, g.Cost(e), e.Cost)
	e = g.EdgeBetween(Node{Id: 3}, Node{Id: 3})
	assert.NotNil(t, e)
	assert.Equal(t, g.Cost(e), e.Cost)
	e = g.EdgeBetween(Node{Id: 3}, Node{Id: 1})
	assert.NotNil(t, e)
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

	g.AddNode(Node{Id: 4})
	assert.Len(t, g.NodeList(), 4)
}

func (suite *GraphTestSuite) TestAddDuplicateNode() {
	t, g := suite.T(), suite.g

	g.AddNode(Node{Id: 4})
	assert.Len(t, g.NodeList(), 4)
	g.AddNode(Node{Id: 4}) // Doing it again shouldn't increase the node count
	assert.Len(t, g.NodeList(), 4)
}

func (suite *GraphTestSuite) TestRemoveNode() {
	t, g := suite.T(), suite.g

	g.RemoveNode(Node{Id: 2})
	assert.Len(t, g.NodeList(), 2)
}

func (suite *GraphTestSuite) TestNonexistentRemoveNode() {
	t, g := suite.T(), suite.g

	g.RemoveNode(Node{Id: 4})
	assert.Len(t, g.NodeList(), 3)
}

func (suite *GraphTestSuite) TestAddDirectedEdge() {
	t, g := suite.T(), suite.g

	g.AddDirectedEdge(&Edge{
		T:    Node{Id: 1},
		H:    Node{Id: 1},
		Cost: 1,
	})
	assert.NotNil(t, g.EdgeBetween(Node{Id: 1}, Node{Id: 1}))
}

func (suite *GraphTestSuite) TestAddDuplicateDirectedEdge() {
	t, g := suite.T(), suite.g

	g.AddDirectedEdge(&Edge{
		T:    Node{Id: 3},
		H:    Node{Id: 3},
		Cost: 1,
	})
	assert.NotNil(t, g.EdgeBetween(Node{Id: 3}, Node{Id: 3}))
}

func (suite *GraphTestSuite) TestRemoveDirectedEdge() {
	t, g := suite.T(), suite.g

	g.RemoveDirectedEdge(&Edge{
		T:    Node{Id: 3},
		H:    Node{Id: 3},
		Cost: 1,
	})
	assert.Nil(t, g.EdgeBetween(Node{Id: 3}, Node{Id: 3}))
}

func (suite *GraphTestSuite) TestRemoveNonexistentDirectedEdge() {
	t, g := suite.T(), suite.g

	g.RemoveDirectedEdge(&Edge{
		T:    Node{Id: 1},
		H:    Node{Id: 1},
		Cost: 1,
	})
	assert.Nil(t, g.EdgeBetween(Node{Id: 1}, Node{Id: 1}))
}

func (suite *GraphTestSuite) TestNodeIsZero() {
	t := suite.T()
	assert.True(t, Node{}.IsZero())
	assert.False(t, Node{Id: 1}.IsZero())
}
