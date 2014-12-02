package route

import (
	"math"
	"sort"
	"time"

	"github.com/volkerp/goquadtree/quadtree"
)

type insertionMode int

const (
	insertionModeBefore insertionMode = iota
	insertionModeBetween
	insertionModeAfter
)

// A RouteInsertion represents a (presumably optimal) point insertion within a route.
type RouteInsertion struct {
	// Route contains the Route into which the point insertion is intended
	Route Route
	// Cost represents the cost (a time duration) of performing the insertion
	Cost time.Duration
	// InsertionPoints are the points between which the insertion is to be performed. If the insertion is to happen at
	// the head or tail of the route, the first or second point will be zero, respectively.
	InsertionPoints [2]Point
	// Point is the point to be inserted
	Point Point
}

// IsHead returns whether the insertion is to be performed at the head of the entire route
func (r RouteInsertion) IsHead() bool {
	return r.InsertionPoints[0].IsZero() && !r.InsertionPoints[1].IsZero()
}

// IsTail returns whether the insertion is to be performed at the tail of the entire route
func (r RouteInsertion) IsTail() bool {
	return r.InsertionPoints[1].IsZero() && !r.InsertionPoints[0].IsZero()
}

// A Route is an immutable representation of a vehicle route between a collection of waypoints.
type Route interface {
	// Bounds returns a pair of bounding co-ordinates (northwest, southeast).
	Bounds() [2]Coordinate
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
	// Insert integrates an insertion defined by the given RouteInsertion (probably derived from InsertionPoints()), and
	// returns a new route
	Insert(RouteInsertion) Route
}

// Represents a point within a route (necessary because a point will be stored in more than one store)
type mappedPoint struct {
	Point        Point
	QuadTreeNode *qtNode
	Idx          int
}

func (m mappedPoint) IsZero() bool {
	return m.QuadTreeNode == nil &&
		m.Idx == 0 &&
		m.Point.IsZero()
}

type routeImpl struct {
	bounds       [2]Coordinate
	coster       Coster
	mappedPoints []mappedPoint
	qt           quadtree.QuadTree
}

func New(coster Coster, points ...Point) Route {
	// Calculate the bounds
	nw := Coordinate{math.Inf(1), math.Inf(-1)}
	se := Coordinate{math.Inf(-1), math.Inf(1)}
	for _, p := range points {
		c := p.Coordinate
		if c[0] < nw[0] {
			nw[0] = c[0]
		}
		if c[1] > nw[1] {
			nw[1] = c[1]
		}
		if c[0] > se[0] {
			se[0] = c[0]
		}
		if c[1] < se[1] {
			se[1] = c[1]
		}
	}

	qt := quadtree.NewQuadTree(quadtree.NewBoundingBox(nw[0], se[0], nw[1], se[1]))
	mappedPoints := make([]mappedPoint, len(points))
	for i, p := range points {
		qtNode := &qtNode{p.Coordinate, i}
		qt.Add(qtNode)

		mappedPoints[i] = mappedPoint{
			Point:        p,
			QuadTreeNode: qtNode,
			Idx:          i,
		}
	}

	return &routeImpl{
		bounds:       [2]Coordinate{nw, se},
		coster:       coster,
		mappedPoints: mappedPoints,
		qt:           qt,
	}
}

func (r *routeImpl) Bounds() [2]Coordinate {
	return r.bounds
}

func (r *routeImpl) Points() []Point {
	result := make([]Point, len(r.mappedPoints))
	for i, mp := range r.mappedPoints {
		result[i] = mp.Point
	}
	return result
}

func (r *routeImpl) Waypoints() []Point {
	result := make([]Point, 0, len(r.mappedPoints)/4)
	for _, mp := range r.mappedPoints {
		if mp.Point.IsWaypoint {
			result = append(result, mp.Point)
		}
	}
	return result
}

func (r *routeImpl) Duration() time.Duration {
	if len(r.mappedPoints) == 0 {
		return time.Duration(0)
	}

	coster := r.coster.Cost
	result := r.mappedPoints[0].Point.Dwell()
	for i := 1; i < len(r.mappedPoints); i++ {
		lp, p := r.mappedPoints[i-1], r.mappedPoints[i]
		result += p.Point.Dwell()
		result += coster(lp.Point.Coordinate, p.Point.Coordinate)
	}
	return result
}

// Returns the best way to insert the given point between the passed existing points, from the given allowable modes
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
		Cost:  cost + p.Dwell(),
		Point: p,
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

func (r *routeImpl) Insert(i RouteInsertion) Route {
	points := r.Points()

	if i.IsHead() { // Head insertion
		points = append(points, Point{})
		copy(points[1:], points[0:])
		points[0] = i.Point
	} else if i.IsTail() { // Tail insertion
		points = append(points, i.Point)
	} else { // Mid insertion
		for ii, candidate := range points {
			if candidate == i.InsertionPoints[0] {
				points = append(points, Point{})
				copy(points[ii+2:], points[ii+1:])
				points[ii+1] = i.Point
				break
			}
		}
	}

	return New(r.coster, points...)
}
