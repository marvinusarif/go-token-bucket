package bucket

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type bucketMock struct {
	b *bucket
}

var (
	ctr                            int64
	emptyMock                      *bucketMock
	mock                           *bucketMock
	errMock                        *bucketMock
	defaultGenerationStrategy      GSFunc
	defaultErrorGenerationStrategy GSFunc
)

const ERR_MADEUP_FOR_TEST = "made up error"

func TestMain(m *testing.M) {
	var newTokens []*Token

	defaultGenerationStrategy = func() (token *Token, err error) {
		atomic.AddInt64(&ctr, 1)
		token = &Token{ShippingID: atomic.LoadInt64(&ctr)}
		return token, err
	}

	defaultErrorGenerationStrategy = func() (*Token, error) {
		return nil, errors.New(ERR_MADEUP_FOR_TEST)
	}
	emptyMock = &bucketMock{b: &bucket{}}

	mock = &bucketMock{
		b: &bucket{
			mu:       &sync.Mutex{},
			capacity: 8,
			run:      defaultGenerationStrategy,
		},
	}
	mock.b.tokens.Store(newTokens)

	errMock = &bucketMock{
		b: &bucket{
			mu:       &sync.Mutex{},
			capacity: 8,
			run:      defaultErrorGenerationStrategy,
		},
	}
	errMock.b.tokens.Store(newTokens)

	os.Exit(m.Run())

}

func TestNewBucket(t *testing.T) {
	actual := NewBucket(8, defaultGenerationStrategy)
	assert.Implements(t, (*Bucket)(nil), actual)
	assert.IsType(t, &bucket{}, actual)
}

func TestLoadStrategy(t *testing.T) {
	emptyMock.b.LoadStrategy(defaultGenerationStrategy)
	assert.NotNil(t, emptyMock.b.run)
}

func TestFill(t *testing.T) {
	tests := []struct {
		name   string
		mock   *bucketMock
		expect error
	}{
		{
			name:   "should return nil",
			mock:   mock,
			expect: nil,
		},
		{
			name:   "should return error",
			mock:   errMock,
			expect: errors.New(ERR_MADEUP_FOR_TEST),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual error
			actual = tt.mock.b.fill()
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestDequeue(t *testing.T) {
	tests := []struct {
		name   string
		args   int
		expect []*Token
	}{
		{
			name: "dequeue 1 token",
			args: 1,
			expect: []*Token{
				&Token{1},
			},
		}, {
			name:   "should return nil when take > remainingTokens (to save unused token)",
			args:   9,
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//reset counter
			atomic.SwapInt64(&ctr, 0)
			mock.b.drain()
			mock.b.fill()

			actual := mock.b.dequeue(tt.args)
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestTake(t *testing.T) {
	tests := []struct {
		name   string
		mock   *bucketMock
		args   []int
		expect []*Token
		err    error
	}{
		{
			name: "take 1 out of (cap)",
			mock: mock,
			args: []int{1},
			expect: []*Token{
				&Token{1},
			},
			err: nil,
		}, {
			name: "take 1 > n <= (cap)",
			mock: mock,
			args: []int{8},
			expect: []*Token{
				&Token{1}, &Token{2},
				&Token{3}, &Token{4},
				&Token{5}, &Token{6},
				&Token{7}, &Token{8},
			},
			err: nil,
		}, {
			name:   "take more than (cap)",
			mock:   mock,
			args:   []int{9},
			expect: nil,
			err:    fmt.Errorf(ERR_TAKE_MORE_THAN_CAP),
		}, {
			name: "take thrice 5-6",
			mock: mock,
			args: []int{5, 6},
			expect: []*Token{
				&Token{6}, &Token{9}, &Token{10},
				&Token{11}, &Token{12}, &Token{13},
			},
			err: nil,
		}, {
			name: "take thrice 5-6-2",
			mock: mock,
			args: []int{5, 6, 2},
			expect: []*Token{
				&Token{7}, &Token{8},
			},
			err: nil,
		}, {
			name:   "error on filling",
			mock:   errMock,
			args:   []int{1},
			expect: nil,
			err:    fmt.Errorf(ERR_MADEUP_FOR_TEST),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//reset all
			atomic.SwapInt64(&ctr, 0)
			tt.mock.b.drain()
			tt.mock.b.fill()
			var (
				actual []*Token
				err    error
			)
			for _, arg := range tt.args {
				actual, err = tt.mock.b.Take(arg)
			}
			assert.ElementsMatch(t, tt.expect, actual)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestLoadToken(t *testing.T) {
	tests := []struct {
		name   string
		expect []*Token
	}{
		{
			name: "load tokens",
			expect: []*Token{
				&Token{1}, &Token{2},
				&Token{3}, &Token{4},
				&Token{5}, &Token{6},
				&Token{7}, &Token{8},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atomic.SwapInt64(&ctr, 0)
			mock.b.drain()
			mock.b.fill()
			actual := mock.b.loadTokens()
			assert.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func TestRemainingToken(t *testing.T) {
	tests := []struct {
		name     string
		take     int
		expected int
	}{
		{
			name:     "fill max cap then take 5",
			take:     5,
			expected: 3,
		}, {
			name:     "fill 8 then take all tokens",
			take:     mock.b.capacity,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atomic.SwapInt64(&ctr, 0)
			mock.b.drain()
			mock.b.fill()
			mock.b.Take(tt.take)
			actual := mock.b.remainingToken()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
