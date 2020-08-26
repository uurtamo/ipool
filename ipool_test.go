package ipool

import (
	"math/rand"
	"testing"
)

func BenchmarkAlloc(b *testing.B) {

	poolio := new(Pool)
	var intArr [10000000]*Interval

	poolio.Init()

	for n := 0; n < 10000000; n++ {
		intArr[n] = poolio.Alloc()
	}

	var result *Interval

	b.ResetTimer()

	for n := 0; n < b.N; n++ {

		result = poolio.Alloc()
		result.begin = rand.Uint64()
		result.end = rand.Uint64()
		poolio.Free(result)
	}
}

func BenchmarkAllocParallel(b *testing.B) {

	poolio := new(Pool)
	var intArr [10000000]*Interval

	poolio.Init()

	for n := 0; n < 10000000; n++ {
		intArr[n] = poolio.Alloc()
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {

		var result *Interval
		for pb.Next() {
			result = poolio.Alloc()
			result.begin = rand.Uint64()
			result.end = rand.Uint64()
			poolio.Free(result)
		}
	})
}

func BenchmarkNew(b *testing.B) {

	var mappy = map[uint64]*Interval{}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {

		yourInterval := new(Interval)
		yourInterval.begin = rand.Uint64()
		yourInterval.end = rand.Uint64()
		yourInterval.handle = rand.Uint64()
		mappy[yourInterval.handle] = yourInterval
	}
}
