package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/obeattie/vrp/graph"
)

func TestAncestors(t *testing.T) {
	suite.Run(t, new(AncestorsTestSuite))
}

type AncestorsTestSuite struct {
	suite.Suite
}

func (suite *AncestorsTestSuite) generateGraph(nodes []nodePrototype, costs float64) graph.Graph {
	g := graph.NewGraph()

	for _, n := range nodes {
		g.AddDirectedEdge(&graph.Edge{
			H:    graph.Node{Id: n.srcId},
			T:    graph.Node{Id: n.targetId},
			Cost: costs,
		})
	}

	return g
}

func (suite *AncestorsTestSuite) TestAncestors() {
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

	ancestors := map[int][]int{
		1: {},
		2: {1},
		3: {1, 2},
		4: {1, 2, 3},
		5: {1, 2, 3, 4},
		6: {1, 2, 3, 4},
		7: {1, 2, 3, 4, 5, 6},
	}

	for srcId, expectedIds := range ancestors {
		returnedAncestors, err := Ancestors(g, graph.Node{Id: srcId})
		assert.NoError(t, err)
		assert.Len(t, returnedAncestors, len(expectedIds))

		// Build a map of each (which can be compared and is not order-sensitive)
		expectedIdsMap := make(map[int]bool, len(expectedIds))
		for _, n := range expectedIds {
			expectedIdsMap[n] = true
		}
		returnedIdsMap := make(map[int]bool, len(returnedAncestors))
		for _, n := range returnedAncestors {
			returnedIdsMap[n.ID()] = true
		}

		assert.Equal(t, expectedIdsMap, returnedIdsMap)
	}
}
