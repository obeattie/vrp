package route

import (
	"time"
)

// Coordinate represents an (x, y) co-ordinate pair. Note that this means the storage format is actually
// (latitude [x], longitude [y]).
type Coordinate [2]float64

func (c *Coordinate) IsZero() bool {
	return c[0] == 0.0 && c[1] == 0.0
}

// A Point represents a place AND a time. That is, a point along a route along with information of WHEN it is visited.
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
