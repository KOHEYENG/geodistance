package main

import (
	"testing"
)

func BenchmarkSphericalTrigonometr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p1 := geoPoint{22.386651, 114.169922}
		p2 := geoPoint{21.4225, 39.8261}

		sphericalTrigonometry(p1, p2)
	}
}

func BenchmarkHubenyFormula(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p1 := geoPoint{22.386651, 114.169922}
		p2 := geoPoint{21.4225, 39.8261}

		hubenyFormula(p1, p2)
	}
}
