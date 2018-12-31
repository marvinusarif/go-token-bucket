package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/token-bucket/bucket"
)

type Token struct {
	ShippingName string
	configID     int
}

//this default function should be used
func (v Token) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintln(&b, v.ShippingName, v.configID)
	return b.Bytes(), nil
}

func (v *Token) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &v.ShippingName, &v.configID)
	return err
}

func main() {
	wg := &sync.WaitGroup{}
	cap := 5
	b := bucket.NewBucket(Token{}, cap, bucket.NewCodec())
	if err := b.AddElem(Token{"A", 1}, 1); err != nil {
		fmt.Println(err)
	}
	if err := b.AddElem(Token{"B", 2}, 1); err != nil {
		fmt.Println(err)
	}
	if err := b.AddElem(Token{"C", 3}, 1); err != nil {
		fmt.Println(err)
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// var x int
			var tokens []Token
			// x = i % cap
			b.Take(&tokens, cap-i%cap)
			// if err != nil {
			// 	fmt.Println(err)
			// }
			// for i := range tokens {
			// 	fmt.Println(i, tokens[i].ShippingName, tokens[i].configID)
			// }
			// if len(tokens) != cap-x {
			// 	fmt.Println("error : expect", cap-x, "actual", len(tokens))
			// }
		}(i)
	}
	wg.Wait()
	for _, i := range b.Occurence() {
		fmt.Printf("%+v\n", i)
	}
	fmt.Println("Done")

}
