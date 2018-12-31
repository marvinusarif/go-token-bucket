package bucket

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/token-bucket/mock"
)

const ERR_MADEUP_FOR_TEST = "made up error"

var (
	mk      *bucket
	codecMk *mocks.Codec
)

func TestMain(m *testing.M) {
	var tokenIndices []int
	elems := []struct {
		name string
		id   int
		prob float64
	}{
		{
			id:   1,
			name: "A",
			prob: 1.0,
		}, {
			id:   2,
			name: "B",
			prob: 1.0,
		}, {
			id:   3,
			name: "C",
			prob: 1.0,
		}, {
			id:   4,
			name: "D",
			prob: 1.0,
		},
	}

	once.Do(func() {
		BufPool = &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}
		gob.Register(DefaultToken{})

	})
	codecMk = new(mocks.Codec)

	mk = &bucket{
		mu:             &sync.Mutex{},
		codec:          NewCodec(),
		capacity:       10,
		elem:           nil,
		totalElem:      0,
		prob:           nil,
		occurence:      nil,
		totalOccurence: 0,
	}

	mk.tokenIndices.Store(tokenIndices)

	for _, e := range elems {
		if err := mk.AddElem(DefaultToken{e.id, e.name}, e.prob); err != nil {
			fmt.Println(err)
		}
	}

	os.Exit(m.Run())
}

func TestAddElem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		arg    interface{}
		arg1   float64
		expect error
	}{
		{
			name: "add 1 element",
			arg:  DefaultToken{1, "A"},
			arg1: 0.1, expect: nil,
		}, {
			name:   "should return err",
			arg:    DefaultToken{0, "A"},
			arg1:   0.1,
			expect: fmt.Errorf(ERR_MADEUP_FOR_TEST),
		},
	}

	mk.codec = codecMk
	codecMk.On("Encode", mock.MatchedBy(func(e interface{}) bool {
		tok := e.(DefaultToken)
		return tok.id == 0
	})).Return(&bytes.Buffer{}, fmt.Errorf(ERR_MADEUP_FOR_TEST))
	codecMk.On("Encode", mock.MatchedBy(func(e interface{}) bool {
		tok := e.(DefaultToken)
		return tok.id != 0
	})).Return(&bytes.Buffer{}, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := mk.AddElem(tt.arg, tt.arg1)
			assert.Equal(t, actual, tt.expect)
		})
	}
}

func TestNewBucket(t *testing.T) {
	t.Parallel()

	actual := NewBucket(DefaultToken{}, mk.capacity, NewCodec())
	assert.Implements(t, (*Bucket)(nil), actual)
	assert.IsType(t, &bucket{}, actual)

	assert.Equal(t, once, once)
}

func TestFill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		totalElement int
		expect       error
	}{
		{
			name:         "should return error due to has no element",
			totalElement: 0,
			expect:       fmt.Errorf(ERR_ELEMENT_MUST_BE_SET),
		},
		{
			name:         "should return nil",
			totalElement: mk.totalElem,
			expect:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual error
			if tt.totalElement != mk.totalElem {
				mk.totalElem = tt.totalElement
			}
			actual = mk.fill()
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestDequeue(t *testing.T) {
	tests := []struct {
		name   string
		args   int
		expect []int
	}{
		{
			name:   "dequeue 1 token index [ 0 1 2 3 0 1 2 3 0 1] => pop 1st element",
			args:   1,
			expect: []int{0},
		}, {
			name:   "should return nil when take > remainingTokens (to save unused token)",
			args:   mk.capacity + 1,
			expect: nil,
		}, {
			name:   "should return 5 token indices",
			args:   6,
			expect: []int{0, 1, 2, 3, 0, 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			//reset
			mk.reset()
			mk.fill()

			actual := mk.dequeue(tt.args)
			assert.ElementsMatch(t, tt.expect, actual)
		})
	}
}

func TestTake(t *testing.T) {
	tests := []struct {
		name                     string
		args                     []int
		errorOnFilling           bool
		errorNotPassingPtr       bool
		errorOnPassingNilOrValue bool
		errorOnNilPtr            bool
		errorOnNotAssignableType bool
		errorOnDecoding          bool
		expect                   []DefaultToken
		err                      error
	}{
		{
			name: "take 1 out of (cap)",
			args: []int{1},
			expect: []DefaultToken{
				DefaultToken{1, "A"},
			},
			err: nil,
		}, {
			name: "take 1 > n <= (cap)",
			args: []int{2, 2},
			expect: []DefaultToken{
				DefaultToken{1, "A"},
				DefaultToken{2, "B"},
				DefaultToken{3, "C"},
				DefaultToken{4, "D"},
			},
			err: nil,
		}, {
			name:   "take more than (cap)",
			args:   []int{mk.capacity + 1},
			expect: nil,
			err:    fmt.Errorf(ERR_TAKE_MORE_THAN_CAP),
		}, {
			name: "take twice 5-6",
			args: []int{5, 6},
			expect: []DefaultToken{
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"}, DefaultToken{4, "D"},
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"}, DefaultToken{4, "D"},
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"},
			},
			err: nil,
		}, {
			name: "take thrice 5-6-5",
			args: []int{5, 6, 5},
			expect: []DefaultToken{
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"}, DefaultToken{4, "D"},
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"}, DefaultToken{4, "D"},
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"}, DefaultToken{4, "D"},
				DefaultToken{1, "A"}, DefaultToken{2, "B"}, DefaultToken{3, "C"}, DefaultToken{4, "D"},
			},
			err: nil,
		}, {
			name:           "error on filling",
			args:           []int{5},
			errorOnFilling: true,
			expect:         nil,
			err:            fmt.Errorf(ERR_ELEMENT_MUST_BE_SET),
		}, {
			name:               "error due to not passing pointer",
			args:               []int{5},
			errorNotPassingPtr: true,
			expect:             nil,
			err:                fmt.Errorf(ERR_MUST_PASS_POINTER),
		}, {
			name:                     "error due to passing nil or value",
			args:                     []int{5},
			errorOnPassingNilOrValue: true,
			expect:                   nil,
			err:                      fmt.Errorf(ERR_MUST_PASS_POINTER),
		}, {
			name:                     "error on not assignable type",
			args:                     []int{5},
			errorOnNotAssignableType: true,
			expect:                   nil,
			err:                      fmt.Errorf(ERR_TYPE_IS_NOT_ASSIGNABLE),
		}, {
			name:          "error due to passing nil",
			args:          []int{5},
			errorOnNilPtr: true,
			expect:        nil,
			err:           fmt.Errorf(ERR_MUST_PASS_POINTER_NOT_NIL),
		}, {
			name:            "error on decoding",
			args:            []int{5},
			errorOnDecoding: true,
			expect:          nil,
			err:             fmt.Errorf(ERR_MADEUP_FOR_TEST),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//reset all
			mk.reset()

			totalElem := mk.totalElem
			if tt.errorOnFilling {
				mk.totalElem = 0
			} else {
				mk.fill()
			}

			var (
				nilPtr *[]DefaultToken
				actual []DefaultToken
				err    error
			)
			for _, v := range tt.args {
				if tt.errorNotPassingPtr {
					err = mk.Take(actual, v)
				} else if tt.errorOnNotAssignableType {
					err = mk.Take(&[]struct{}{}, v)
				} else if tt.errorOnPassingNilOrValue {
					err = mk.Take(nil, v)
				} else if tt.errorOnNilPtr {
					err = mk.Take(nilPtr, v)
				} else if tt.errorOnDecoding {
					mk.codec = codecMk
					codecMk.On("Decode", mock.Anything, mock.Anything).Return(fmt.Errorf(ERR_MADEUP_FOR_TEST))
					err = mk.Take(&actual, v)
				} else {
					err = mk.Take(&actual, v)
				}
			}
			assert.ElementsMatch(t, tt.expect, actual)
			assert.Equal(t, tt.err, err)

			//reset to initial condition
			mk.totalElem = totalElem
		})
	}
}

func TestOccurence(t *testing.T) {
	tests := []struct {
		name     string
		expected []int64
	}{
		{
			name:     "return occurence",
			expected: mk.occurence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mk.fill()
			actual := mk.Occurence()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
