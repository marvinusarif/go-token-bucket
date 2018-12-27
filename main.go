package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/token-bucket/bucket"
)

func main() {
	wg := &sync.WaitGroup{}
	var counter int64

	generationStrategy := func() (token *bucket.Token, err error) {
		ctr := atomic.LoadInt64(&counter)
		atomic.AddInt64(&counter, 1)
		token = &bucket.Token{ShippingID: ctr}
		return token, err
	}

	b := bucket.NewBucket(250, generationStrategy)
	for i := 0; i < 200; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()
			var x int
			x = i % 8
			if x == 0 {
				x = 1
			}
			tokens, err := b.Take(8 - x)
			if err != nil {
				fmt.Println(err)
			}
			for i := range tokens {
				fmt.Println(tokens[i])
			}
			if len(tokens) != 8-x {
				fmt.Println("error : expect", 8-x, "actual", len(tokens))
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("Done")

}
