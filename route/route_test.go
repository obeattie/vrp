package route

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRoute(t *testing.T) {
	suite.Run(t, new(RouteTestSuite))
}

type RouteTestSuite struct {
	suite.Suite
	r Route
}

func (suite *RouteTestSuite) SetupTest() {
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
	suite.r = New(HaversineCoster{}, points...)
}

func (suite *RouteTestSuite) TestInsertionPoints() {
	t, r := suite.T(), suite.r

	expectations := map[Coordinate][2]string{
		Coordinate{-43.1882863, -22.9116324}:      {"", "Home"},
		Coordinate{-0.13152, 51.42581}:            {"", "Home"},
		Coordinate{-0.16573906, 51.45636018}:      {"Home", "Clapham Junction"},
		Coordinate{-0.1664257, 51.47042378}:       {"Clapham Junction", "Soho Square"},
		Coordinate{-0.1123051, 51.5031653}:        {"Somerset House", ""},
		Coordinate{18.0685808, 59.32932349999999}: {"Somerset House", ""},
	}

	for newCoord, between := range expectations {
		result := r.InsertionPoints(Point{
			Key:        "Test",
			IsWaypoint: true,
			Coordinate: newCoord,
		})
		predecessor, successor := result[0], result[1]

		if between[0] != predecessor.Key || between[1] != successor.Key {
			assert.Fail(t, fmt.Sprintf("%v should have inserted at %v, got %v", newCoord, between,
				[2]string{predecessor.Key, successor.Key}))
		}
	}
}
