// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import bytes "bytes"
import mock "github.com/stretchr/testify/mock"

// Codec is an autogenerated mock type for the Codec type
type Codec struct {
	mock.Mock
}

// Decode provides a mock function with given fields: b, e
func (_m *Codec) Decode(b []byte, e interface{}) error {
	ret := _m.Called(b, e)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, interface{}) error); ok {
		r0 = rf(b, e)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Encode provides a mock function with given fields: e
func (_m *Codec) Encode(e interface{}) (*bytes.Buffer, error) {
	ret := _m.Called(e)

	var r0 *bytes.Buffer
	if rf, ok := ret.Get(0).(func(interface{}) *bytes.Buffer); ok {
		r0 = rf(e)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bytes.Buffer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(e)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}