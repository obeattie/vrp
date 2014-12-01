package route

import (
	"sort"
)

// RouteInsertions is a sortable collection of InsertionResults
type RouteInsertions []RouteInsertion

func (i RouteInsertions) Len() int {
	return len(i)
}

func (i RouteInsertions) Less(x, y int) bool {
	return i[x].Cost < i[y].Cost
}

func (i RouteInsertions) Swap(x, y int) {
	i[x], i[y] = i[y], i[x]
}

func computeInsertion(r Route, p Point, c chan<- RouteInsertion) {
	c <- r.InsertionPoints(p)
}

// FindClosestRoutes locates the cheapest routes in which to insert the given Point.
func FindClosestRoutes(p Point, candidates []Route, n int) RouteInsertions {
	if n > len(candidates) {
		n = len(candidates)
	}

	resultC := make(chan RouteInsertion, 10)
	results := make(RouteInsertions, 0, n+1)

	for _, candidate := range candidates {
		go computeInsertion(candidate, p, resultC)
	}
	for i := 0; i < len(candidates); i++ {
		results = append(results, <-resultC)
		sort.Sort(results)
		if len(results) > n {
			results = results[:n]
		}
	}

	return results
}
