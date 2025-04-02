package algorithms

func WaterFilling(demands []int64, capacity int64) []int64 {
	capacityRemaining := capacity
	output := make([]int64, len(demands))

	for capacityRemaining > 0 {
		for i, demand := range demands {
			if output[i] < demand {
				output[i]++
				capacityRemaining--
				if capacityRemaining == 0 {
					break
				}
			}
		}

		allSatisfied := true
		for i, demand := range demands {
			if output[i] < demand {
				allSatisfied = false
				break
			}
		}
		if allSatisfied {
			break
		}
	}

	return output
}
