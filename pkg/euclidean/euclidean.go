package euclidean

import (
	"math"

	"github.com/8ff/dhf64/pkg/misc"
)

func EuclideanDistance(min, max float64, constraint int, data []float64) []int {
	output := make([]int, len(data))
	// Generate all possible codewords for constaint - 1 bits, and add a xor parity bit
	combos := misc.GenerateCodewords(constraint - 1)
	for r := 0; r < len(data); r += constraint { // Go over all the points
		bestCombo := make([]int, constraint)
		bestDist := math.MaxFloat64
		for combo := 0; combo < len(combos); combo++ { // Go over all combos of 1/0s
			distance := 0.0
			for c := 0; c < constraint; c++ { // Generate distances for n points
				// Print out individual value of formula
				distance += math.Pow((float64(combos[combo][c]) - data[r+c]), 2)
				// fmt.Printf("r: %f, combo: %f c: %d distance: %f\n", data[r+c], float64(combos[combo][c]), c, distance)
			}
			// Print out total distance
			// fmt.Printf("Total distance: %f\n", distance)
			if distance < bestDist {
				bestDist = distance
				bestCombo = combos[combo]
			}
		}
		// Print best distance and combo
		// fmt.Printf("Best distance: %f, best combo: %v\n", bestDist, bestCombo)
		// Store bestCombo in output
		for i := 0; i < constraint; i++ {
			output[r+i] = bestCombo[i]
		}

	}

	return output
}
