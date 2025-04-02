package algorithms

func FairShare(demands []int64, capacity int64) []int64 {
	output := make([]int64, len(demands))
	capacityRemaining := capacity

	for capacityRemaining > 0 {
		// Count how many users still need more resources
		activeUsers := 0
		for i, demand := range demands {
			if output[i] < demand {
				activeUsers++
			}
		}

		// If all demands are met, break
		if activeUsers == 0 {
			break
		}

		// Calculate fair share per active user
		fairShare := capacityRemaining / int64(activeUsers)
		if fairShare == 0 {
			fairShare = 1
		}

		// Allocate fair share to each active user
		for i, demand := range demands {
			if output[i] < demand {
				allocate := fairShare
				if output[i]+allocate > demand {
					allocate = demand - output[i]
				}
				output[i] += allocate
				capacityRemaining -= allocate

				if capacityRemaining <= 0 {
					break
				}
			}
		}
	}

	return output
}
