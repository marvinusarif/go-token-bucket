// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Bucket is an autogenerated mock type for the Bucket type
type Bucket struct {
	mock.Mock
}

// AddElem provides a mock function with given fields: e, p
func (_m *Bucket) AddElem(e interface{}, p float64) error {
	ret := _m.Called(e, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, float64) error); ok {
		r0 = rf(e, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Occurence provides a mock function with given fields:
func (_m *Bucket) Occurence() []int64 {
	ret := _m.Called()

	var r0 []int64
	if rf, ok := ret.Get(0).(func() []int64); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	return r0
}

// Take provides a mock function with given fields: t, n
func (_m *Bucket) Take(t interface{}, n int) error {
	ret := _m.Called(t, n)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, int) error); ok {
		r0 = rf(t, n)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}