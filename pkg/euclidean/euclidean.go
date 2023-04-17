package euclidean

import (
	"fmt"
	"math"

	"github.com/8ff/udarp/pkg/misc"
)

func HybridDecodeV4(data []float64) [][]int {
	// Very memory intensive, for test set generated 400k records with best match of 3 errors
	mean := 0.0
	sum := 0.0
	negativeMeanTotal := 0.0
	positiveMeanTotal := 0.0
	negativeMean := 0.0
	positiveMean := 0.0
	negativeCount := 0
	positiveCount := 0

	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	for _, v := range data {
		if v < mean {
			negativeMeanTotal += math.Abs(mean - v)
			negativeCount++
		} else {
			positiveMeanTotal += math.Abs(v - mean)
			positiveCount++
		}
	}

	negativeMean = negativeMeanTotal / float64(negativeCount)
	positiveMean = positiveMeanTotal / float64(positiveCount)

	// Preallocate variations
	variations := make([][]int, 1)
	variations[0] = make([]int, len(data))

	for i, v := range data {
		newVariations := make([][]int, 0)

		for _, variation := range variations {
			if v < mean {
				if v > mean-negativeMean {
					newVariation := make([]int, len(variation))
					copy(newVariation, variation)
					newVariation[i] = 1
					newVariations = append(newVariations, newVariation)
				}
				variation[i] = 0
			} else {
				if v < mean+positiveMean {
					newVariation := make([]int, len(variation))
					copy(newVariation, variation)
					newVariation[i] = 0
					newVariations = append(newVariations, newVariation)
				}
				variation[i] = 1
			}
		}

		variations = append(variations, newVariations...)
	}

	return variations
}

func HybridDecodeV4V1(data []float64) [][]int {
	/* Plan
	Split range by mean call it 0.5
	Go over data and find LO/Hi mean of Abs(mean-data[i]) call those 0.25 and 0.75
	Go over data and and store variations of 0 and 1
	Pass 1
	If data < mean 0.5 = 0
	If data > mean 0.5 = 1
	Pass 2
	If data > mean 0.25 = 1
	If data < mean 0.75 = 0
	*/

	variations := make([][]int, 1)
	variations[0] = make([]int, len(data))

	sum := 0.0
	mean := 0.0
	positiveMeanTotal := 0.0
	negativeMeanTotal := 0.0
	// positiveMean := 0.0
	negativeMean := 0.0

	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			negativeMeanTotal += math.Abs(mean - data[i])
			negativeMean = negativeMeanTotal / float64(i+1)
		} else {
			positiveMeanTotal += math.Abs(data[i] - mean)
			// positiveMean = positiveMeanTotal / float64(i+1)
		}
	}

	newVariations := make([][]int, 0)
	for i, v := range data {
		for _, variation := range variations {
			if v < mean {
				// distanceToMean := math.Abs(v - mean)
				// distanceToZero := math.Abs(v - negativeMean)
				// distanceNegativeMean := math.Abs(v - negativeMean)
				// fmt.Printf("%d > 0: v: %f m: %f d2m: %f nm: %f d2r: %f\n", i+1, data[i], mean, mean-data[i], negativeMean, distanceNegativeMean)
				if v > negativeMean {
					// Maybe a 1
					// fmt.Printf("Maybe a 1 ?\n")
					newVariation := make([]int, len(variation))
					copy(newVariation, variation)
					newVariation[i] = 1
					newVariations = append(newVariations, newVariation)
				}
				variation[i] = 0
			} else {
				// distancePositiveMean := math.Abs(v - positiveMean)
				// fmt.Printf("%d > 1: v: %f m: %f d2m: %f pm: %f d2r: %f\n", i+1, data[i], mean, data[i]-mean, positiveMean, distancePositiveMean)
				// if v < positiveMean {
				// 	// Maybe a 0
				// 	// fmt.Printf("Maybe a 0 ?\n")
				// 	newVariation := make([]int, len(variation))
				// 	copy(newVariation, variation)
				// 	newVariation[i] = 0
				// 	newVariations = append(newVariations, newVariation)
				// }
				variation[i] = 1
			}
		}
		variations = append(variations, newVariations...)
		newVariations = newVariations[:0]
	}

	return variations
}

func HardDecode(data []float64) []int {
	sum := 0.0
	mean := 0.0

	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	fmt.Printf("Mean: %f\n", mean)
	output := make([]int, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			output[i] = 0
		} else {
			output[i] = 1
		}
	}
	return output
}

// func HybridDecodeV2(data []float64) []int {
// 	// TODO: Notes to implement
// 	// If distance to mean equal to mean, then its a solid 0 or 1
// 	// If distance to mean is 0, then its a maybe 0 or 1

// 	// for i := 0; i < len(data); i++ {
// 	// 	data[i] = data[i] * 100000
// 	// }

// 	sum := 0.0
// 	mean := 0.0
// 	positiveMeanTotal := 0.0
// 	negativeMeanTotal := 0.0
// 	positiveMean := 0.0
// 	negativeMean := 0.0

// 	for _, v := range data {
// 		sum += v
// 	}
// 	mean = sum / float64(len(data))

// 	// Calculate postive and negative means
// 	for i := 0; i < len(data); i++ {
// 		if data[i] < mean {
// 			negativeMeanTotal += math.Abs(mean - data[i])
// 			negativeMean = negativeMeanTotal / float64(i)
// 		} else {
// 			positiveMeanTotal += math.Abs(data[i] - mean)
// 			positiveMean = positiveMeanTotal / float64(i)
// 		}
// 	}

// 	output := make([]int, len(data))
// 	for i := 0; i < len(data); i++ {
// 		if data[i] < mean {
// 			distanceNegativeMean := math.Abs(data[i] - negativeMean)
// 			if distanceNegativeMean < negativeMean {
// 				output[i] = 1
// 				continue
// 			}
// 			output[i] = 0
// 		} else {

// 			distancePositiveMean := math.Abs(data[i] - positiveMean)
// 			if distancePositiveMean < positiveMean {
// 				output[i] = 0
// 				continue
// 			}
// 			output[i] = 1
// 		}
// 	}
// 	return output
// }

func HardDecodeV2Debug(data []float64) []int {
	// TODO: Goal, is to return a [][]int with all possible decodes
	// TODO: Notes to implement
	// If distance to mean equal to mean, then its a solid 0 or 1
	// If distance to mean is 0, then its a maybe 0 or 1

	for i := 0; i < len(data); i++ {
		data[i] = data[i] * 100000
	}

	sum := 0.0
	mean := 0.0
	positiveMeanTotal := 0.0
	negativeMeanTotal := 0.0
	positiveMean := 0.0
	negativeMean := 0.0

	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	// Calculate postive and negative means
	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			negativeMeanTotal += math.Abs(mean - data[i])
			negativeMean = negativeMeanTotal / float64(i)
			// fmt.Printf("v: %f m: %f d: %f r: %f\n", data[i], mean, mean-data[i], negativeMeanTotal/float64(i))
		} else {
			positiveMeanTotal += math.Abs(data[i] - mean)
			positiveMean = positiveMeanTotal / float64(i)
			// fmt.Printf("v: %f m: %f d: %f r: %f\n", data[i], mean, data[i]-mean, positiveMeanTotal/float64(i))
		}
	}

	fmt.Printf("Mean: %f Positive Mean: %f Negative Mean: %f\n", mean, positiveMean, negativeMean)
	output := make([]int, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			distanceNegativeMean := math.Abs(data[i] - negativeMean)
			fmt.Printf("%d > 0: v: %f m: %f d2m: %f nm: %f d2r: %f\n", i+1, data[i], mean, mean-data[i], negativeMean, distanceNegativeMean)
			if distanceNegativeMean < negativeMean {
				fmt.Printf("Maybe a 1 ?\n")
				output[i] = 1
				continue
			}
			output[i] = 0
		} else {

			distancePositiveMean := math.Abs(data[i] - positiveMean)
			fmt.Printf("%d > 1: v: %f m: %f d2m: %f pm: %f d2r: %f\n", i+1, data[i], mean, data[i]-mean, positiveMean, distancePositiveMean)
			if distancePositiveMean < positiveMean {
				fmt.Printf("Maybe a 0 ?\n")
				output[i] = 0
				continue
			}
			output[i] = 1
		}
	}
	return output
}

func HybridDecodeV3Debug(data []float64) [][]int {
	variations := make([][]int, 1)
	variations[0] = make([]int, len(data))

	sum := 0.0
	mean := 0.0
	positiveMeanTotal := 0.0
	negativeMeanTotal := 0.0
	positiveMean := 0.0
	negativeMean := 0.0

	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			negativeMeanTotal += math.Abs(mean - data[i])
			negativeMean = negativeMeanTotal / float64(i+1)
		} else {
			positiveMeanTotal += math.Abs(data[i] - mean)
			positiveMean = positiveMeanTotal / float64(i+1)
		}
	}

	// TODO: Maybe try this simple logic, split 0-1 in 4 parts, and generate combos based on that

	newVariations := make([][]int, 0)
	for i, v := range data {
		for _, variation := range variations {
			if v < mean {
				distanceToMean := math.Abs(v - mean)
				distanceToZero := math.Abs(v - negativeMean)
				distanceNegativeMean := math.Abs(v - negativeMean)
				fmt.Printf("%d > 0: v: %f m: %f d2m: %f nm: %f d2r: %f\n", i+1, data[i], mean, mean-data[i], negativeMean, distanceNegativeMean)
				if distanceNegativeMean < negativeMean {
					// Maybe a 1
					fmt.Printf("Maybe a 1 ?\n")
					newVariation := make([]int, len(variation))
					copy(newVariation, variation)
					newVariation[i] = 1
					newVariations = append(newVariations, newVariation)
				}
				// If distance to mean is 0, then its a maybe 0 or 1
				if distanceToMean < distanceToZero {
					// Maybe a 1
					fmt.Printf("Maybe a 1 ?\n")
					newVariation := make([]int, len(variation))
					copy(newVariation, variation)
					newVariation[i] = 1
					newVariations = append(newVariations, newVariation)
				}
				variation[i] = 0
			} else {
				distancePositiveMean := math.Abs(v - positiveMean)
				fmt.Printf("%d > 1: v: %f m: %f d2m: %f pm: %f d2r: %f\n", i+1, data[i], mean, data[i]-mean, positiveMean, distancePositiveMean)
				if distancePositiveMean < positiveMean {
					// Maybe a 0
					fmt.Printf("Maybe a 0 ?\n")
					newVariation := make([]int, len(variation))
					copy(newVariation, variation)
					newVariation[i] = 0
					newVariations = append(newVariations, newVariation)
				}
				variation[i] = 1
			}
		}
		variations = append(variations, newVariations...)
		newVariations = newVariations[:0]
	}

	return variations
}

func HybridDecodeV3DebugOld(data []float64) [][]int {
	// As we go over the data, if there is a maybe, clone each [x][]int and add a 0 and 1 to each and store them in a new array of [x][]int
	// TODO: Goal, is to return a [][]int with all possible decodes
	// TODO: Notes to implement
	// If distance to mean equal to mean, then its a solid 0 or 1
	// If distance to mean is 0, then its a maybe 0 or 1

	variations := make([][]int, 0)
	// Pre declare the first variation
	variations = append(variations, make([]int, len(data)))

	// for i := 0; i < len(data); i++ {
	// 	data[i] = data[i] * 100000
	// }

	sum := 0.0
	mean := 0.0
	positiveMeanTotal := 0.0
	negativeMeanTotal := 0.0
	positiveMean := 0.0
	negativeMean := 0.0

	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	// Calculate postive and negative means
	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			negativeMeanTotal += math.Abs(mean - data[i])
			negativeMean = negativeMeanTotal / float64(i)
			// fmt.Printf("v: %f m: %f d: %f r: %f\n", data[i], mean, mean-data[i], negativeMeanTotal/float64(i))
		} else {
			positiveMeanTotal += math.Abs(data[i] - mean)
			positiveMean = positiveMeanTotal / float64(i)
			// fmt.Printf("v: %f m: %f d: %f r: %f\n", data[i], mean, data[i]-mean, positiveMeanTotal/float64(i))
		}
	}

	fmt.Printf("Mean: %f Positive Mean: %f Negative Mean: %f\n", mean, positiveMean, negativeMean)
	for i := 0; i < len(data); i++ {
		if data[i] < mean {
			distanceNegativeMean := math.Abs(data[i] - negativeMean)
			fmt.Printf("%d > 0: v: %f m: %f d2m: %f nm: %f d2r: %f\n", i+1, data[i], mean, mean-data[i], negativeMean, distanceNegativeMean)
			if distanceNegativeMean < negativeMean {
				fmt.Printf("Maybe a 1 ?\n")
				variations[0][i] = 0
				// output[i] = 1
				// continue
			}
			// variations[0][i] = 0
		} else {

			distancePositiveMean := math.Abs(data[i] - positiveMean)
			fmt.Printf("%d > 1: v: %f m: %f d2m: %f pm: %f d2r: %f\n", i+1, data[i], mean, data[i]-mean, positiveMean, distancePositiveMean)
			if distancePositiveMean < positiveMean {
				fmt.Printf("Maybe a 0 ?\n")
				// output[i] = 0
				// continue
			}
			// variations[0][i] = 1
		}
	}

	return variations
}

// This is a wrapper for the EuclideanDistance function
func SoftDecode(constraint int, data []float64) ([]int, error) {
	if len(data) < constraint {
		return nil, fmt.Errorf("constraint must be less than or equal to length of data")
	}
	output := make([]int, 0)

	for i := 0; i < len(data); i++ {
		if i+constraint > len(data) {
			break
		}

		bits := EuclideanDistance(constraint, data[i:i+constraint])
		if i == 0 {
			// Keep all bits for first run
			output = append(output, bits...)
			// fmt.Printf("First run: %d\n", bits)
		} else {
			// Append last bit to output
			output = append(output, bits[len(bits)-1])
			// fmt.Printf("Subsequent run: %d\n", bits[len(bits)-1])
		}

	}
	return output, nil
}

// This works poorly
func EuclideanDistance(constraint int, data []float64) []int {
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

// func GenerateCombinations(base []int, markers []int) [][]int {
// 	var markerCount = 0
// 	var maxLimit = 1
// 	// Count how many 1s are in the markers
// 	for i, marker := range markers {
// 		if marker == 1 {
// 			markerCount++
// 			if markerCount > maxLimit {
// 				markers[i] = 0
// 			}
// 		}
// 	}
// 	fmt.Printf("Marker count: %d\n", markerCount)
// 	fmt.Printf("Markers: %v\n", markers)

// 	combinations := make([][]int, 0)

// 	// Calculate the number of combinations to generate
// 	combinationCount := 1 << len(markers)

// 	for i := 0; i < combinationCount; i++ {
// 		combination := make([]int, len(base))
// 		copy(combination, base)

// 		for j, marker := range markers {
// 			if marker == 1 {
// 				// Toggle bit at jth position
// 				combination[j] = (combination[j] + (i>>j)&1) % 2
// 			}
// 		}

// 		// Check if the combination already exists in the result
// 		exists := false
// 		for _, existingCombination := range combinations {
// 			match := true
// 			for k, v := range existingCombination {
// 				if v != combination[k] {
// 					match = false
// 					break
// 				}
// 			}
// 			if match {
// 				exists = true
// 				break
// 			}
// 		}

// 		if !exists {
// 			combinations = append(combinations, combination)
// 		}
// 	}

// 	return combinations
// }
