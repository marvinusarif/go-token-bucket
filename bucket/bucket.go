package bucket

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

const (
	ERR_TAKE_MORE_THAN_CAP        = "trying to grab token more than its capacity"
	ERR_MUST_PASS_POINTER         = "must pass a pointer, not a value"
	ERR_MUST_PASS_POINTER_NOT_NIL = "nil pointer passed to destination"
	ERR_ELEMENT_MUST_BE_SET       = "bucket should have at least 1 element"
	ERR_TYPE_IS_NOT_ASSIGNABLE    = "type is not assignable"
)

var (
	BufPool *sync.Pool
	once    *sync.Once = &sync.Once{}
)

type Bucket interface {
	Take(t interface{}, n int) error
	AddElem(e interface{}, p float64) error
	Occurence() []int64
}

type bucket struct {
	mu             *sync.Mutex
	codec          Codec
	elem           [][]byte
	totalElem      int
	prob           []float64
	totalProb      float64
	occurence      []int64
	totalOccurence int
	capacity       int
	tokenIndices   atomic.Value
}

func NewBucket(tokenType interface{}, cap int, codec Codec) Bucket {
	once.Do(func() {
		BufPool = &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}
	})

	gob.Register(tokenType)

	newBucket := &bucket{
		mu:             &sync.Mutex{},
		codec:          codec,
		capacity:       cap,
		elem:           nil,
		totalElem:      0,
		prob:           nil,
		occurence:      nil,
		totalOccurence: 0,
	}

	var tokenIndices []int
	newBucket.tokenIndices.Store(tokenIndices)
	return newBucket
}

func (b *bucket) AddElem(e interface{}, p float64) (err error) {
	buf, err := b.codec.Encode(e)
	if err != nil {
		return err
	}

	b.elem = append(b.elem, buf.Bytes())
	b.prob = append(b.prob, p)
	b.totalElem++
	b.totalProb += p
	b.occurence = append(b.occurence, 0)

	return nil
}

func (b *bucket) loadTokens() []int {
	return b.tokenIndices.Load().([]int)
}

func (b *bucket) remainingToken() int {
	return len(b.loadTokens())
}

func (b *bucket) Dereference(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (b *bucket) Take(t interface{}, n int) (err error) {
	tokenIndices, err := b.take(n)
	if err != nil {
		return err
	}

	valuePtr := reflect.ValueOf(t)
	if valuePtr.Kind() != reflect.Ptr {
		return fmt.Errorf(ERR_MUST_PASS_POINTER)
	}

	if valuePtr.IsNil() {
		return fmt.Errorf(ERR_MUST_PASS_POINTER_NOT_NIL)
	}

	value := valuePtr.Elem()
	baseType := b.Dereference(valuePtr.Type())
	baseElem := baseType.Elem()
	//instantiate new object based on the type of base element
	e := reflect.New(baseElem).Elem().Interface()

	// fmt.Println(e)
	// fmt.Println("base element type :", baseElem)
	// fmt.Println("total elements :", len(b.elem))
	// fmt.Println("token indices:", tokenIndices)

	for _, j := range tokenIndices {

		if err = b.codec.Decode(b.elem[j], &e); err != nil {
			return err
		}

		if isAssignable := baseElem.AssignableTo(reflect.TypeOf(e)); !isAssignable {
			return fmt.Errorf(ERR_TYPE_IS_NOT_ASSIGNABLE)
		}

		value.Set(reflect.Append(value, reflect.ValueOf(e)))
	}
	return nil
}

func (b *bucket) take(n int) (tokensIndices []int, err error) {
	if n > b.capacity {
		return tokensIndices, fmt.Errorf(ERR_TAKE_MORE_THAN_CAP)
	} else {
		tokensIndices = b.dequeue(n)
	}

	if len(tokensIndices) < n {
		if err := b.fill(); err != nil {
			return tokensIndices, err
		}
		return b.take(n)
	}
	return tokensIndices, nil
}

func (b *bucket) generateToken() (tokenIndex int) {
	return int(b.totalOccurence % b.totalElem)
}

func (b *bucket) fill() (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var (
		tokensIndices []int
		newTokenIndex int
		left          int
	)

	if b.totalElem < 1 {
		return fmt.Errorf(ERR_ELEMENT_MUST_BE_SET)
	}

	left = b.capacity - b.remainingToken()
	for i := 0; i < left; i++ {
		newTokenIndex = b.generateToken()
		tokensIndices = append(tokensIndices, newTokenIndex)
		b.occurence[newTokenIndex]++
		b.totalOccurence++
	}
	b.put(tokensIndices[0:])
	return err
}

func (b *bucket) Occurence() []int64 {
	return b.occurence
}

func (b *bucket) dequeue(n int) (tokens []int) {
	remainingTokenIndices := b.loadTokens()
	if n > len(remainingTokenIndices) {
		//return when number of token is less than expect value
		return tokens
	}
	tokens, remainingTokenIndices = remainingTokenIndices[0:n], remainingTokenIndices[n:]
	b.tokenIndices.Store(remainingTokenIndices)
	return tokens
}

func (b *bucket) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.totalOccurence = 0
	b.drain()
}
func (b *bucket) drain() {
	var newTokenIndices []int
	b.tokenIndices.Store(newTokenIndices)
}

func (b *bucket) put(tokens []int) {
	var newTokenIndices []int
	newTokenIndices = append(newTokenIndices, b.loadTokens()[0:]...)
	newTokenIndices = append(newTokenIndices, tokens...)
	b.tokenIndices.Store(newTokenIndices)
}
