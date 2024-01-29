package accounting

import "math"

func splitCostEvenly(cost int, n int) []int {
	total := 0
	chunk := int(math.Floor(float64(cost) / float64(n)))
	chunks := make([]int, n)
	for i := 0; i < n-1; i++ {
		total += chunk
		chunks[i] = chunk
	}
	// put the remainder in the last chunk
	chunks[n-1] = cost - total
	return chunks
}
