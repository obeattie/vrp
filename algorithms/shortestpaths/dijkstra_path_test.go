package shortestpaths

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/obeattie/vrp/graph"
)

type nodePrototype struct {
	srcId, targetId int
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
			H:    graph.Node{Id: n.srcId},
			T:    graph.Node{Id: n.targetId},
			Cost: costs,
		})
	}

	return g
}

func (suite *DijkstraPathTestSuite) idFromNodeOrId(id interface{}) int {
	switch id := id.(type) {
	case int:
		return id
	case graph.Node:
		return id.ID()
	default:
		assert.Fail(suite.T(), "idFromNodeOrId must be passed int or node")
		return -1
	}
}

func (suite *DijkstraPathTestSuite) idsFromNodesOrIds(ids interface{}) []int {
	switch ids := ids.(type) {
	case []int:
		return ids
	case []graph.Node:
		result := make([]int, len(ids))
		for i, node := range ids {
			result[i] = suite.idFromNodeOrId(node)
		}
		return result
	default:
		assert.Fail(suite.T(), "idsFromNodesOrIds must be passed ints or nodes")
		return nil
	}
}

func (suite *DijkstraPathTestSuite) pathMatches(_p1, _p2 interface{}) bool {
	p1, p2 := suite.idsFromNodesOrIds(_p1), suite.idsFromNodesOrIds(_p2)

	if len(p1) != len(p2) {
		return false
	}
	for i, n := range p1 {
		n1, n2 := suite.idFromNodeOrId(n), suite.idFromNodeOrId(p2[i])
		if n1 != n2 {
			return false
		}
	}
	return true
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

	paths, costs, err := singleSourceDijkstra(g, graph.Node{Id: 1}, graph.Node{}, math.Inf(0))
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
		for _, candidatePath := range validPaths[destination.ID()] {
			if suite.pathMatches(candidatePath, path) {
				matched = true
				break destinationLoop
			}
		}
	}

	assert.True(t, matched)

	for destination, cost := range costs {
		assert.Equal(t, (len(paths[destination])-1)*2, cost)
	}
}

func (suite *DijkstraPathTestSuite) TestDijkstraPath() {
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

	shortestPaths := map[[2]int][][]int{
		[2]int{1, 1}: {{1}},
		[2]int{1, 2}: {{1, 2}},
		[2]int{1, 3}: {{1, 2, 3}},
		[2]int{1, 4}: {{1, 2, 3, 4}},
		[2]int{1, 5}: {{1, 2, 3, 4, 5}},
		[2]int{1, 6}: {{1, 2, 3, 4, 6}},
		[2]int{1, 7}: {{1, 2, 3, 4, 5, 7},
			{1, 2, 3, 4, 6, 7}},
		[2]int{4, 7}: {{4, 5, 7},
			{4, 6, 7}},
	}
	for _origindest, validPaths := range shortestPaths {
		origin, dest := _origindest[0], _origindest[1]
		matched := false
	candidatePathLoop:
		for _, candidatePath := range validPaths {
			returnedPath, err := DijkstraPath(g, graph.Node{Id: origin}, graph.Node{Id: dest})
			assert.NoError(t, err, "Error retrieving path")
			if suite.pathMatches(candidatePath, returnedPath) {
				matched = true
				break candidatePathLoop
			}
			t.Logf("%+v", returnedPath)
		}
		assert.True(t, matched, fmt.Sprintf("Valid path from %d -> %d not returned", origin, dest))
	}
}
