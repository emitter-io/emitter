/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/

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
