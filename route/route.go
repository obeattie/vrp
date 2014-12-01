package route

import (
	"math"
	"sort"
	"time"

	"github.com/volkerp/goquadtree/quadtree"

	"github.com/obeattie/vrp/graph"
)

type insertionMode int

const (
	insertionModeBefore insertionMode = iota
	insertionModeBetween
	insertionModeAfter
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

type RouteInsertion struct {
	Route           Route
	Cost            time.Duration
	InsertionPoints [2]Point
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
	// InsertionPoints return the points between which a given Point should be optimally inserted (at lowest cost), and
	// the cost of doing so. If it is most optimal to insert at the head or tail of the entire route, the first or
	// second result will be zero, respectively.
	InsertionPoints(p Point) RouteInsertion
	// KNearest returns the k-nearest points to a given Coordinate
	KNearest(c Coordinate, k int) []Point
	// Equal returns whether the passed routes are equivalent
	Equal(r Route) bool
}

type mappedPoint struct {
	GraphNode    graph.Node
	Point        Point
	QuadTreeNode *qtNode
	Idx          int
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
			Idx:          i,
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

// Returns the best way to insert the given point between the passed existing points, from the given allowable insertion
// modes
func (r *routeImpl) optimalLegInsertion(leg []mappedPoint, p Point, modes ...insertionMode) (insertionMode, time.Duration) {
	coster := r.coster.Cost
	originalCost := time.Duration(0)

	if len(modes) == 0 || len(leg) == 0 {
		return insertionModeAfter, time.Duration(math.MaxInt64)
	} else if len(leg) == 1 {
		return modes[0], coster(leg[0].Point.Coordinate, p.Coordinate)
	}

	mode, costDiff := modes[0], time.Duration(math.MaxInt64)
	for i := 1; i < len(leg); i++ {
		seg1, seg2 := leg[i-1].Point, leg[i].Point
		originalSegCost := coster(seg1.Coordinate, seg2.Coordinate)
		originalCost += originalSegCost

		for _, candidateMode := range modes {
			modeCostDiff := time.Duration(math.MaxInt64)
			switch candidateMode {
			case insertionModeBefore:
				modeCostDiff = coster(p.Coordinate, seg1.Coordinate)
			case insertionModeBetween:
				modeCost := coster(seg1.Coordinate, p.Coordinate) + coster(p.Coordinate, seg2.Coordinate)
				modeCostDiff = modeCost - originalSegCost
			case insertionModeAfter:
				modeCostDiff = coster(seg2.Coordinate, p.Coordinate)
			}
			if modeCostDiff < costDiff {
				mode, costDiff = candidateMode, modeCostDiff
			}
		}
	}

	return mode, costDiff
}

// Returns tuples (predecessor, p) (p, successor), if available
func (r *routeImpl) legTuples(p mappedPoint) [][]mappedPoint {
	result := make([][]mappedPoint, 0, 2)
	if p.Idx != 0 {
		result = append(result, []mappedPoint{r.mappedPoints[p.Idx-1], p})
	}
	if p.Idx < len(r.mappedPoints)-1 {
		result = append(result, []mappedPoint{p, r.mappedPoints[p.Idx+1]})
	}
	if len(result) < 1 {
		result = append(result, []mappedPoint{p})
	}
	return result
}

func (r *routeImpl) InsertionPoints(p Point) RouteInsertion {
	candidates := r.mappedPoints

	insertionLeg, insertionMode, cost := []mappedPoint{}, insertionModeAfter, time.Duration(math.MaxInt64)
	for _, tuple := range r.legTuples(candidates[0]) { // Head
		if legMode, legCost := r.optimalLegInsertion(tuple, p, insertionModeBefore); legCost < cost {
			insertionLeg, insertionMode, cost = tuple, legMode, legCost
		}
	}
	nearestIdx := r.kNearestMappedPointIndices(p.Coordinate, 1)[0]
	for _, tuple := range r.legTuples(candidates[nearestIdx]) { // Surrounding the nearest node
		if legMode, legCost := r.optimalLegInsertion(tuple, p, insertionModeBetween); legCost < cost {
			insertionLeg, insertionMode, cost = tuple, legMode, legCost
		}
	}
	for _, tuple := range r.legTuples(candidates[len(candidates)-1]) { // Tail
		if legMode, legCost := r.optimalLegInsertion(tuple, p, insertionModeAfter); legCost < cost {
			insertionLeg, insertionMode, cost = tuple, legMode, legCost
		}
	}

	result := RouteInsertion{
		Route: r,
		Cost:  cost,
	}

	switch insertionMode {
	case insertionModeBefore:
		result.InsertionPoints = [2]Point{{}, insertionLeg[0].Point}
	case insertionModeBetween:
		result.InsertionPoints = [2]Point{insertionLeg[0].Point, insertionLeg[1].Point}
	case insertionModeAfter:
		result.InsertionPoints = [2]Point{insertionLeg[1].Point}
	}

	return result
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

func (r *routeImpl) Equal(other Route) bool {
	otherPoints := other.Points()
	if len(otherPoints) != len(r.mappedPoints) {
		return false
	}

	for i, p := range otherPoints {
		if p != r.mappedPoints[i].Point {
			return false
		}
	}

	return true
}
