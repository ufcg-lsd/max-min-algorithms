package main

import (
	"fmt"
	"maxmin/algorithms"
)

func main() {
	demands := []int64{4, 6, 8}
	capacity := int64(13)

	fmt.Println("Water Filling:", algorithms.WaterFilling(demands, capacity))
	fmt.Println("Fair Share:", algorithms.FairShare(demands, capacity))
	fmt.Println("Max-Min LSD:", algorithms.MaxMin(demands, capacity))
}
