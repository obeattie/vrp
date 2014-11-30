package route

import (
	"math"
	"sort"
	"time"

	"github.com/volkerp/goquadtree/quadtree"

	"github.com/obeattie/vrp/graph"
)

// Coordinate represents an (x, y) co-ordinate pair. Note that this means the storage format is actually
// (latitude [x], longitude [y]).
type Coordinate [2]float64

func (c *Coordinate) IsZero() bool {
	return c[0] == 0.0 && c[1] == 0.0
}

// Point represents a place AND a time. That is, a point along a route along with information of WHEN it is visited.
// A point may be either a waypoint (a node which representing some stop along the route), or a route point
// (representing the route between the waypoints).
type Point struct {
	Arrival    time.Time
	Coordinate Coordinate
	Departure  time.Time
	IsWaypoint bool
	Key        string
}

// Dwell returns a duration of the dwell time at the
func (p Point) Dwell() time.Duration {
	return p.Departure.Sub(p.Arrival)
}

func (p Point) IsZero() bool {
	return p.IsWaypoint == false &&
		p.Key == "" &&
		p.Arrival.IsZero() &&
		p.Departure.IsZero() &&
		p.Coordinate.IsZero()
}

// A Route is an immutable representation of a vehicle route between a collection of waypoints.
type Route interface {
	// Bounds returns a pair of bounding co-ordinates (northwest, southeast).
	Bounds() [2]Coordinate
	// Graph returns a Graph object representing all known points (waypoints and routing points) as vertices with
	// time-costed edges (edge costs are estimated nanoseconds of travel time). Changing the graph will NOT update
	// the route.
	Graph() graph.Graph
	// Points returns an ordered collection of points in the route.
	Points() []Point
	// Waypoints returns an ordered collection of waypoints in the route.
	Waypoints() []Point
	// Duration returns the total duration of the route, including dwell time at points.
	Duration() time.Duration
	// InsertionPoints return the points between which a given Point should be optimally inserted (at lowest cost).
	InsertionPoints(p Point) [2]Point
	// KNearest returns the k-nearest points to a given Coordinate
	KNearest(c Coordinate, k int) []Point
}

type mappedPoint struct {
	GraphNode    graph.Node
	Point        Point
	QuadTreeNode *qtNode
}

func (m mappedPoint) IsZero() bool {
	return m.GraphNode.IsZero() && m.Point.IsZero()
}

type routeImpl struct {
	bounds       [2]Coordinate
	coster       Coster
	graph        graph.Graph
	mappedPoints []mappedPoint
	duration     time.Duration
	qt           quadtree.QuadTree
}

func New(coster Coster, points ...Point) Route {
	// Calculate the bounds
	nw := Coordinate{math.Inf(1), math.Inf(1)}
	se := Coordinate{math.Inf(-1), math.Inf(-1)}
	for _, p := range points {
		c := p.Coordinate
		if c[0] < nw[0] {
			nw[0] = c[0]
		}
		if c[1] < nw[1] {
			nw[1] = c[1]
		}
		if c[0] > se[0] {
			se[0] = c[0]
		}
		if c[1] > se[1] {
			se[1] = c[1]
		}
	}

	qt := quadtree.NewQuadTree(quadtree.NewBoundingBox(nw[0], se[0], nw[1], se[1]))
	g := graph.NewGraph()
	mappedPoints := make([]mappedPoint, len(points))
	duration := time.Duration(0) // Calculate this ahead of tie for faster retrieval

	for i, p := range points { // Add graph nodes
		duration += p.Dwell()

		node := g.NewNode()
		node.Lat = p.Coordinate[1]
		node.Lng = p.Coordinate[0]
		g.AddNode(node)

		qtNode := &qtNode{p.Coordinate, i}
		qt.Add(qtNode)

		mappedPoints[i] = mappedPoint{
			GraphNode:    node,
			Point:        p,
			QuadTreeNode: qtNode,
		}
	}
	for i, mp := range mappedPoints { // Add graph edges
		if i == 0 {
			continue
		}
		lastMp := mappedPoints[i-1]
		cost := coster.Cost(lastMp.Point.Coordinate, mp.Point.Coordinate)
		duration += cost
		g.AddDirectedEdge(&graph.Edge{
			H:    lastMp.GraphNode,
			T:    mp.GraphNode,
			Cost: float64(cost.Nanoseconds()),
		})
	}

	return &routeImpl{
		bounds:       [2]Coordinate{nw, se},
		coster:       coster,
		graph:        g,
		mappedPoints: mappedPoints,
		duration:     duration,
		qt:           qt,
	}
}

func (r *routeImpl) Bounds() [2]Coordinate {
	return r.Bounds()
}

func (r *routeImpl) Graph() graph.Graph {
	return r.graph.Copy() // We do not want mutations to this affecting the Route
}

func (r *routeImpl) Points() []Point {
	result := make([]Point, len(r.mappedPoints))
	for i, mp := range r.mappedPoints {
		result[i] = mp.Point
	}
	return result
}

func (r *routeImpl) Waypoints() []Point {
	result := make([]Point, len(r.mappedPoints)/4)
	for _, mp := range r.mappedPoints {
		if mp.Point.IsWaypoint {
			result = append(result, mp.Point)
		}
	}
	return result
}

func (r *routeImpl) Duration() time.Duration {
	return r.duration
}

// nearestPoints returns a pair of Points which are nearest to the given Point (which may or may not be in the Route).
// The returned points are (predecessor, successor). If the point given is nearest only to the first or last point
// already in the route, then the predecessor or successor may be zero.
func (r *routeImpl) nearestPoints(p Point) [2]Point {
	cost := r.coster.Cost
	candidates := r.mappedPoints
	bestIdx := r.kNearestMappedPointIndices(p.Coordinate, 1)[0]
	best := candidates[bestIdx]
	bestCost := cost(best.Point.Coordinate, p.Coordinate)

	result := [2]Point{}
	// Is the new point nearer to the best point's predecessor, or its successor?
	predecessor, successor := mappedPoint{}, mappedPoint{}
	predecessorCost, successorCost := time.Duration(math.MaxInt64), time.Duration(math.MaxInt64)
	unalteredPredecessorCost, unalteredSuccessorCost := time.Duration(0), time.Duration(0)
	if bestIdx > 0 {
		predecessor = candidates[bestIdx-1]
		predecessorCost = bestCost + cost(predecessor.Point.Coordinate, p.Coordinate)
		unalteredPredecessorCost = cost(predecessor.Point.Coordinate, best.Point.Coordinate)
	}
	if bestIdx >= 0 && bestIdx < len(candidates)-1 {
		successor = candidates[bestIdx+1]
		successorCost = bestCost + cost(p.Coordinate, successor.Point.Coordinate)
		unalteredSuccessorCost = cost(best.Point.Coordinate, successor.Point.Coordinate)
	}

	// Because we care about what the overall effect on the (predecessor, best, successor) segment of the route will be,
	// include the unaltered segment's cost in an option's cost
	predecessorCost += unalteredSuccessorCost
	successorCost += unalteredPredecessorCost

	if predecessorCost < successorCost { // Ties go to the successor
		if bestIdx == len(candidates)-1 { // Consider inserting as the route destination
			afterCost := bestCost + cost(predecessor.Point.Coordinate, best.Point.Coordinate)
			betweenCost := bestCost + cost(predecessor.Point.Coordinate, p.Coordinate)
			if afterCost < betweenCost {
				result[0] = best.Point
				return result
			}
		}
		result[0], result[1] = predecessor.Point, best.Point
	} else {
		if bestIdx == 0 { // Consider inserting as the route origin
			beforeCost := bestCost + cost(best.Point.Coordinate, successor.Point.Coordinate)
			betweenCost := bestCost + cost(p.Coordinate, successor.Point.Coordinate)
			if beforeCost < betweenCost {
				result[1] = best.Point
				return result
			}
		}
		result[0], result[1] = best.Point, successor.Point
	}

	return result
}

func (r *routeImpl) InsertionPoints(p Point) [2]Point {
	return r.nearestPoints(p)
}

func (r *routeImpl) kNearestMappedPointIndices(c Coordinate, k int) []int {
	candidates := r.mappedPoints
	if k > len(candidates) {
		k = len(candidates)
	}

	results := make([]int, 0, k)

	for i, radius := 0, 0.0; len(results) < cap(results); i, radius = i+1, math.Pow(500.0, float64(i)) {
		// We need to sort the results before inserting, to ensure we are actually getting the k-nearest
		sorter := &qtNodeSorter{
			nodes:  r.qt.Query(qtBbox(c, radius)),
			origin: c,
			coster: r.coster,
		}
		sort.Sort(sorter)
		for _, result := range sorter.nodes {
			results = append(results, result.(*qtNode).i)
			if len(results) >= cap(results) {
				break
			}
		}
	}

	return results
}

func (r *routeImpl) KNearest(c Coordinate, k int) []Point {
	indices := r.kNearestMappedPointIndices(c, k)
	results := make([]Point, len(indices))
	for i, ii := range indices {
		results[i] = r.mappedPoints[ii].Point
	}
	return results
}
