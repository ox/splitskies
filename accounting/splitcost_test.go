package accounting

import (
	"reflect"
	"testing"
)

func TestSplitCostEvenly(t *testing.T) {
	tests := []struct {
		cost     int
		n        int
		expected []int
	}{
		{100, 3, []int{33, 33, 34}},
		{50, 2, []int{25, 25}},
		{5, 2, []int{2, 3}},
		{1_000_000_000, 3, []int{333_333_333, 333_333_333, 333_333_334}},
	}

	for i, costTest := range tests {
		result := splitCostEvenly(costTest.cost, costTest.n)
		if !reflect.DeepEqual(costTest.expected, result) {
			t.Errorf("test %d, expected: %v, got %v", i, costTest.expected, result)
		}
	}
}
