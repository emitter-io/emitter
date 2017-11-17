package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMock(t *testing.T) {

	c := NewMockClient()

	url := "testurl.com"
	buf := []byte("test")
	out := new(testObject)

	// Get(url string, output interface{}, headers ...HeaderValue) error
	c.On("Get", url, mock.Anything, mock.Anything).Return(nil).Once()
	assert.Nil(t, c.Get(url, out))

	// Post(url string, body []byte, output interface{}, headers ...HeaderValue) error
	c.On("Post", url, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	assert.Nil(t, c.Post(url, buf, out))

	// PostJSON(url string, body interface{}, output interface{}) (err error)
	c.On("PostJSON", url, mock.Anything, mock.Anything).Return(nil).Once()
	assert.Nil(t, c.PostJSON(url, out, out))

	// PostBinary(url string, body interface{}, output interface{}) (err error)
	c.On("PostBinary", url, mock.Anything, mock.Anything).Return(nil).Once()
	assert.Nil(t, c.PostBinary(url, out, out))
}
