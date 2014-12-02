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
		originalDuration := r.Duration()

		result := r.InsertionPoints(Point{
			Key:        "Test",
			IsWaypoint: true,
			Coordinate: newCoord,
		})
		assert.True(t, result.Cost > 0, "Cost expected to be greater than 0")
		predecessor, successor := result.InsertionPoints[0], result.InsertionPoints[1]

		if between[0] != predecessor.Key || between[1] != successor.Key {
			assert.Fail(t, fmt.Sprintf("%v should have inserted at %v, got %v (%s)", newCoord, between,
				[2]string{predecessor.Key, successor.Key}, result.Cost.String()))
		}

		newRoute := r.Insert(result)
		newDuration := newRoute.Duration()
		actualCost := newDuration - originalDuration
		assert.Equal(t, result.Cost, actualCost, fmt.Sprintf("%s (actual) != %s (expected)", actualCost.String(),
			result.Cost.String()))
	}
}

func (suite *RouteTestSuite) TestKNearest() {
	t, r := suite.T(), suite.r

	result := r.KNearest(Coordinate{-43.1882863, -22.9116324}, 1)
	assert.Len(t, result, 1)
	assert.Equal(t, "Home", result[0].Key)

	result = r.KNearest(Coordinate{-43.1882863, -22.9116324}, 2)
	assert.Len(t, result, 2)
	assert.Equal(t, "Home", result[0].Key)
}

func (suite *RouteTestSuite) TestInsert() {
	t, r := suite.T(), suite.r
	ps := r.Points()
	newP := Point{
		Coordinate: Coordinate{-0.098234, 51.376165},
		Key:        "NEW",
	}

	// Head
	ins := RouteInsertion{
		Route:           r,
		InsertionPoints: [2]Point{{}, ps[0]},
		Point:           newP,
	}
	newR := r.Insert(ins)
	expectedPs := append([]Point{newP}, ps...)
	assert.Len(t, expectedPs, len(ps)+1)
	assert.Equal(t, expectedPs, newR.Points())

	// Middle
	ins.InsertionPoints = [2]Point{ps[0], ps[1]}
	newR = r.Insert(ins)
	expectedPs = append([]Point{ps[0], newP}, ps[1:]...)
	assert.Len(t, expectedPs, len(ps)+1)
	assert.Equal(t, expectedPs, newR.Points())

	// Tail
	ins.InsertionPoints = [2]Point{ps[len(ps)-1], {}}
	newR = r.Insert(ins)
	expectedPs = append(ps, newP)
	assert.Len(t, expectedPs, len(ps)+1)
	assert.Equal(t, expectedPs, newR.Points())
}
