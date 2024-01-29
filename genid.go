package main

import mathrand "math/rand"

var allowedIDCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func generateID(n int) string {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		ret[i] = allowedIDCharacters[mathrand.Intn(36)]
	}
	return string(ret)
}
