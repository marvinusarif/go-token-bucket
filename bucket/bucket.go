package bucket

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	ERR_TAKE_MORE_THAN_CAP = "trying to grab token more than its capacity"
)

type GSFunc func() (*Token, error)

type Bucket interface {
	Take(n int) ([]*Token, error)
	LoadStrategy(gsf GSFunc)
}

type bucket struct {
	mu       *sync.Mutex
	capacity int
	tokens   atomic.Value
	run      GSFunc
}

func NewBucket(cap int, gsf GSFunc) Bucket {
	var newTokens []*Token
	newBucket := &bucket{
		mu:       &sync.Mutex{},
		capacity: cap,
		run:      gsf,
	}
	newBucket.tokens.Store(newTokens)
	return newBucket
}

func (b *bucket) LoadStrategy(gsf GSFunc) {
	b.run = gsf
}

func (b *bucket) loadTokens() []*Token {
	return b.tokens.Load().([]*Token)
}

func (b *bucket) remainingToken() int {
	return len(b.loadTokens())
}

func (b *bucket) Take(n int) (tokens []*Token, err error) {
	if n > b.capacity {
		return tokens, fmt.Errorf(ERR_TAKE_MORE_THAN_CAP)
	} else {
		tokens = b.dequeue(n)
	}
	if len(tokens) < n {
		if err := b.fill(); err != nil {
			return tokens, err
		}
		return b.Take(n)
	}
	return tokens, nil
}

func (b *bucket) fill() (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var (
		tokens   []*Token
		newToken *Token
	)
	for i := 0; i < b.capacity-b.remainingToken(); i++ {
		newToken, err = b.run()
		if err != nil {
			return err
		}
		tokens = append(tokens, newToken)
	}
	b.put(tokens[0:])
	return err
}

func (b *bucket) dequeue(n int) (tokens []*Token) {
	remainingTokens := b.loadTokens()
	if n > len(remainingTokens) {
		//return when number of token is less than expect value
		return tokens
	}
	tokens, rTokens := remainingTokens[0:n], remainingTokens[n:]
	b.tokens.Store(rTokens)
	return tokens
}

func (b *bucket) drain() {
	var emptyTokens []*Token
	b.tokens.Store(emptyTokens)
}

func (b *bucket) put(tokens []*Token) {
	var newTokens []*Token
	newTokens = append(newTokens, tokens...)
	newTokens = append(newTokens, b.loadTokens()[0:]...)
	b.tokens.Store(newTokens)
}
