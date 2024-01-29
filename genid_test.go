package main

import (
	"testing"
)

func TestGenID(t *testing.T) {
	id := generateID(8)
	if len(id) != 8 {
		t.Fail()
	}

	// for i := 0; i < 10; i++ {
	// 	fmt.Printf("%s, ", generateID(8))
	// }
	// t.Fail()
}
