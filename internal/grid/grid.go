package grid

import (
	"fmt"
	"math"
)

const (
	MilesPerDegreeLat = 69.0
	MilesPerDegreeLng = 54.6 // approximate at UK latitudes (~52-55N)
)

func CellID(lat, lng float64) string {
	y := int(math.Floor(lat * MilesPerDegreeLat))
	x := int(math.Floor(lng * MilesPerDegreeLng))
	return fmt.Sprintf("%d_%d", y, x)
}

func AdjacentCells(cellID string) []string {
	var cy, cx int
	fmt.Sscanf(cellID, "%d_%d", &cy, &cx)

	cells := make([]string, 0, 9)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cells = append(cells, fmt.Sprintf("%d_%d", cy+dy, cx+dx))
		}
	}
	return cells
}

func CellsInRadius(cellID string, radius int) []string {
	var cy, cx int
	fmt.Sscanf(cellID, "%d_%d", &cy, &cx)

	cells := make([]string, 0, (2*radius+1)*(2*radius+1))
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			cells = append(cells, fmt.Sprintf("%d_%d", cy+dy, cx+dx))
		}
	}
	return cells
}
