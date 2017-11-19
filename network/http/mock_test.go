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
	c.On("Get", url, mock.Anything, mock.Anything).Return([]byte{}, nil).Once()
	_, e1 := c.Get(url, out)
	assert.Nil(t, e1)

	// Post(url string, body []byte, output interface{}, headers ...HeaderValue) error
	c.On("Post", url, mock.Anything, mock.Anything, mock.Anything).Return([]byte{}, nil).Once()
	_, e2 := c.Post(url, buf, out)
	assert.Nil(t, e2)

}
