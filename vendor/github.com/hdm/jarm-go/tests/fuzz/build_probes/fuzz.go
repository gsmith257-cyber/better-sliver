package fuzz

import "github.com/hdm/jarm-go"

// Fuzz uses go-fuzz to test BuildProbe()
func Fuzz(data []byte) int {
	fuzzProbes := jarm.GetProbes(string(data), 443)
	for _, probe := range fuzzProbes {
		jarm.BuildProbe(probe)
	}
	return 1
}
