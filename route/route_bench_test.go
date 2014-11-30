package route

import (
	"testing"
)

func setupBench() (Route, Point) {
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
	actualPoints := make([]Point, 0, 5000)
	for i := 0; i < cap(actualPoints)/len(points); i++ {
		actualPoints = append(actualPoints, points...)
	}

	r := New(HaversineCoster{}, actualPoints...)
	insertion := Point{
		Coordinate: Coordinate{-0.16573906, 51.45636018},
		IsWaypoint: true,
	}

	return r, insertion
}

func BenchmarkRouteInsertionPoints(b *testing.B) {
	r, insertion := setupBench()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.InsertionPoints(insertion)
	}
}

func BenchmarkRouteInsertionPointsParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		r, insertion := setupBench()
		for pb.Next() {
			r.InsertionPoints(insertion)
		}
	})
}
