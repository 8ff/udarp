package main

import (
	"fmt"

	"github.com/8ff/udarp/pkg/euclidean"
)

func main() {
	var data = []float64{0.2, 0.4, 0.7}
	constraint := 3
	res := euclidean.EuclideanDistance(0.0, 0.9, constraint, data)
	fmt.Println(res)
}
