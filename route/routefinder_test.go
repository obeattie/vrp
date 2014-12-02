package route

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var routeTestFiles = [...]string{
	"testdata/0.json", // 0: Brighton -> Cambridge
	"testdata/1.json", // 1: Inverness -> Mousehole
	"testdata/2.json", // 2: Richmond -> Chiswick -> Baker St -> Elephant & Castle
	"testdata/3.json", // 3: Croydon -> Haringey
}

func TestRouteFinder(t *testing.T) {
	suite.Run(t, new(RouteFinderTestSuite))
}

type RouteFinderTestSuite struct {
	suite.Suite
	routes []Route
}

func (suite *RouteFinderTestSuite) SetupTest() {
	t := suite.T()
	nodes := 0

	suite.routes = make([]Route, len(routeTestFiles))
	for i, path := range routeTestFiles {
		data, err := ioutil.ReadFile(path)
		assert.NoError(t, err)

		var geojson map[string]interface{}
		err = json.Unmarshal(data, &geojson)

		assert.NoError(t, err)

		points := make([]Point, 0, 500)
		for ii, _p := range geojson["geometry"].(map[string]interface{})["coordinates"].([]interface{}) {
			p := _p.([]interface{})
			points = append(points, Point{
				Coordinate: Coordinate{p[0].(float64), p[1].(float64)},
				Key:        strconv.Itoa(ii),
			})
			nodes++
		}
		suite.routes[i] = New(HaversineCoster, points...)
	}

	t.Logf("Has %d vertices", nodes)
}

func (suite *RouteFinderTestSuite) routeIndex(r Route) int {
	for i, candidate := range suite.routes {
		if r.Equal(candidate) {
			return i
		}
	}
	return -1
}

func (suite *RouteFinderTestSuite) TestFindClosestRoutes() {
	t := suite.T()

	expectations := map[Coordinate][]int{
		Coordinate{-0.113468, 51.553807}: {3, 2, 0, 1}, // Holloway Rd
		Coordinate{-5.051041, 50.263195}: {1, 0, 2, 3}, // Truro
		Coordinate{-3.188267, 55.953252}: {1, 0, 3, 2}, // Edinburgh
		Coordinate{51.239208, -0.16988}:  {0, 3, 2, 1}, // Redhill
	}

	for c, expectedIndices := range expectations {
		p := Point{
			Coordinate: c,
			IsWaypoint: true,
		}
		results := FindClosestRoutes(p, suite.routes, len(suite.routes))
		assert.Len(t, results, len(expectedIndices))

		for i, r := range results {
			actualI := suite.routeIndex(r.Route)
			pass := assert.Equal(t, expectedIndices[i], actualI, strconv.Itoa(i))
			if !pass {
				t.Logf("Route %d (cost: %s)", actualI, r.Cost.String())
				t.Logf("Nearest points: [%v %v]", r.InsertionPoints[0].Coordinate, r.InsertionPoints[1].Coordinate)
			}
		}
	}
}
