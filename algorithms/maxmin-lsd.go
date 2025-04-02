package algorithms

func MaxMin(demands []int64, capacity int64) []int64 {
	capacityRemaining := capacity
	output := make([]int64, len(demands))

	for i, demand := range demands {
		share := capacityRemaining / int64(len(demands)-i)
		allocation := min(share, demand)

		if i == len(demands)-1 {
			if demand >= capacityRemaining {
				allocation = capacityRemaining
			}
		}

		output[i] = allocation
		capacityRemaining -= allocation
	}

	if capacityRemaining != 0 {
		for i := range output {
			add := capacityRemaining / int64(len(output))
			output[i] = output[i] + add
		}
	}

	return output
}
