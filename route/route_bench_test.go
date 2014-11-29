package route

import (
	"testing"
)

func BenchmarkRouteInsertionPoints(b *testing.B) {
	points := []Point{
		{
			Key:        "Home",
			IsWaypoint: true,
			Coordinate: Coordinate{-0.1555536, 51.4323465},
		},
		{
			Key:        "Clapham Junction",
			IsWaypoint: true,
			Coordinate: Coordinate{-0.17027, 51.46418999999999},
		},
		{
			Key:        "Soho Square",
			IsWaypoint: true,
			Coordinate: Coordinate{-0.1321499, 51.51530770000001},
		},
		{
			Key:        "Somerset House",
			IsWaypoint: true,
			Coordinate: Coordinate{-0.1174437, 51.510761},
		},
	}
	r := New(HaversineCoster{}, points...)
	insertion := Point{
		Coordinate: Coordinate{-0.16573906, 51.45636018},
		IsWaypoint: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.InsertionPoints(insertion)
	}
}
