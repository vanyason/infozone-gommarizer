package main

import "testing"

func init() {
	if err := CreateTestData(); err != nil {
		panic(err)
	}
}

func BenchmarkSingleThread(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TestSingleThread()
	}
}

func Benchmark1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Test1000()
	}
}

func BenchmarkWP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TestWP()
	}
}
