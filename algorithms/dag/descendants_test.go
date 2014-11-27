package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/obeattie/vrp/graph"
)

func TestDescendants(t *testing.T) {
	suite.Run(t, new(DescendantsTestSuite))
}

type DescendantsTestSuite struct {
	suite.Suite
}

func (suite *DescendantsTestSuite) generateGraph(nodes []nodePrototype, costs float64) graph.Graph {
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

func (suite *DescendantsTestSuite) TestDescendants() {
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

	descendants := map[int][]int{
		1: {2, 3, 4, 5, 6, 7},
		2: {3, 4, 5, 6, 7},
		3: {4, 5, 6, 7},
		4: {5, 6, 7},
		5: {7},
		6: {7},
		7: {},
	}

	for srcId, expectedIds := range descendants {
		returneddescendants, err := Descendants(g, graph.Node{Id: srcId})
		assert.NoError(t, err)
		assert.Len(t, returneddescendants, len(expectedIds))

		// Build a map of each (which can be compared and is not order-sensitive)
		expectedIdsMap := make(map[int]bool, len(expectedIds))
		for _, n := range expectedIds {
			expectedIdsMap[n] = true
		}
		returnedIdsMap := make(map[int]bool, len(returneddescendants))
		for _, n := range returneddescendants {
			returnedIdsMap[n.ID()] = true
		}

		assert.Equal(t, expectedIdsMap, returnedIdsMap)
	}
}
