// Code generated by mockery v1.1.2. DO NOT EDIT.

package survey

import (
	"github.com/emitter-io/emitter/internal/event"
	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/emitter/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/weaveworks/mesh"
)

// pubsubMock is an autogenerated mock type for the broker type
type pubsubMock struct {
	mock.Mock
}

// Publish provides a mock function with given fields: _a0, _a1
func (_m *pubsubMock) Publish(_a0 *message.Message, _a1 func(message.Subscriber) bool) int64 {
	ret := _m.Called(_a0, _a1)

	var r0 int64
	if rf, ok := ret.Get(0).(func(*message.Message, func(message.Subscriber) bool) int64); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// Subscribe provides a mock function with given fields: _a0, _a1
func (_m *pubsubMock) Subscribe(_a0 message.Subscriber, _a1 *event.Subscription) bool {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	if rf, ok := ret.Get(0).(func(message.Subscriber, *event.Subscription) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields: _a0, _a1
func (_m *pubsubMock) Unsubscribe(_a0 message.Subscriber, _a1 *event.Subscription) bool {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	if rf, ok := ret.Get(0).(func(message.Subscriber, *event.Subscription) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields: _a0, _a1
func (_m *pubsubMock) Handle(_ string, _ service.Handler) {
	panic("not implemented")
}

// ------------------------------------------------------------------------------------

// gossiper is an autogenerated mock type for the gossiper type
type gossiperMock struct {
	mock.Mock
}

// ID provides a mock function with given fields:
func (_m *gossiperMock) ID() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// NumPeers provides a mock function with given fields:
func (_m *gossiperMock) NumPeers() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// SendTo provides a mock function with given fields: _a0, _a1
func (_m *gossiperMock) SendTo(_a0 mesh.PeerName, _a1 *message.Message) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(mesh.PeerName, *message.Message) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ------------------------------------------------------------------------------------

// surveyeeMock is an autogenerated mock type for the Surveyee type
type surveyeeMock struct {
	mock.Mock
}

// OnSurvey provides a mock function with given fields: queryType, request
func (_m *surveyeeMock) OnSurvey(queryType string, request []byte) ([]byte, bool) {
	ret := _m.Called(queryType, request)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, []byte) []byte); ok {
		r0 = rf(queryType, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string, []byte) bool); ok {
		r1 = rf(queryType, request)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}
