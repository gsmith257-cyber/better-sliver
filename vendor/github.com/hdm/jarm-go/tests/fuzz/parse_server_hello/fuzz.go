package fuzz

import "github.com/hdm/jarm-go"

var fuzzProbes = jarm.GetProbes("placeholder", 443)

// Fuzz uses go-fuzz to test ParseServerHello()
func Fuzz(data []byte) int {
	for _, probe := range fuzzProbes {
		jarm.ParseServerHello(data, probe)
	}
	return 1
}
