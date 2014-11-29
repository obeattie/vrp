package route

import (
	"time"

	hgeo "github.com/hailocab/go-hailo-lib/geo"
)

const vehicleSpeed = 15.0 * 1000.0 / (60.0 * 60.0 * 1000.0) // Meters per millisecond (15kph)

type Coster interface {
	Cost(c1, c2 Coordinate) time.Duration
}

// HaversineCoster calculates costs based on a Haversine distance and an approximated vehicle speed (4kph).
type HaversineCoster struct{}

func (c HaversineCoster) Cost(c1, c2 Coordinate) time.Duration {
	meters := hgeo.HaversineInMeters(c1[1], c1[0], c2[1], c2[0])
	return time.Duration(meters/vehicleSpeed) * time.Millisecond
}
