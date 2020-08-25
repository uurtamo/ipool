package ipool

import (
	//	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkAlloc(b *testing.B) {

	//	var intList [100]*interval

	//	myInterval := new(interval)
	poolio := new(pool)
	var intArr [10000000]*interval

	//	myInterval.begin = rand.Uint64()
	//	myInterval.end = rand.Uint64()

	poolio.Init()

	for n := 0; n < 10000000; n++ {
		intArr[n] = poolio.Alloc()
	}

	var result *interval

	b.ResetTimer()

	for n := 0; n < b.N; n++ {

		result = poolio.Alloc()
		result.begin = rand.Uint64()
		result.end = rand.Uint64()
		poolio.Free(result)
	}
}

func BenchmarkAllocParallel(b *testing.B) {

	//	var intList [100]*interval

	//	myInterval := new(interval)
	poolio := new(pool)
	var intArr [10000000]*interval

	//	myInterval.begin = rand.Uint64()
	//	myInterval.end = rand.Uint64()

	poolio.Init()

	for n := 0; n < 10000000; n++ {
		intArr[n] = poolio.Alloc()
	}

	//	for n := 0; n < 10000000; n++ {
	//		poolio.Free(intArr[n])
	//	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {

		var result *interval
		for pb.Next() {
			result = poolio.Alloc()
			result.begin = rand.Uint64()
			result.end = rand.Uint64()
			poolio.Free(result)
		}
	})
}

func BenchmarkNew(b *testing.B) {

	//	var intList [100]*interval

	//	myInterval := new(interval)
	var mappy = map[uint64]*interval{}

	//	myInterval.begin = rand.Uint64()
	//	myInterval.end = rand.Uint64()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {

		yourInterval := new(interval)
		yourInterval.begin = rand.Uint64()
		yourInterval.end = rand.Uint64()
		yourInterval.handle = rand.Uint64()
		mappy[yourInterval.handle] = yourInterval
	}
}
