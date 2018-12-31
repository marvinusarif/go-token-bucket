package bucket_test

import (
	"testing"

	"github.com/token-bucket/bucket"
)

func BenchmarkTokenBucket(b *testing.B) {
	cap := 10
	mk := bucket.NewBucket(bucket.DefaultToken{}, cap, bucket.NewCodec())

	benchmarks := []struct {
		name  string
		Token interface{}
		take  []int
	}{
		{"take 1 token", []bucket.DefaultToken{}, []int{1}},
		{"take 5 token", []bucket.DefaultToken{}, []int{5}},
		{"take 8,5,7,5 token", []bucket.DefaultToken{}, []int{8, 5, 7, 5}},
	}

	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				for _, t := range benchmark.take {
					mk.Take(benchmark.Token, t)
				}
			}
		})
	}
}
