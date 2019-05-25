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
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of Client
type MockClient struct {
	mock.Mock
}

// NewMockClient returns a mock implementation of Client
func NewMockClient() *MockClient {
	return &MockClient{
		Mock: mock.Mock{},
	}
}

// Get issues an HTTP Get on a specified URL and decodes the payload as JSON.
func (mock *MockClient) Get(url string, output interface{}, headers ...HeaderValue) ([]byte, error) {
	mockArgs := mock.Called(url, output, headers)
	return mockArgs.Get(0).([]byte), mockArgs.Error(1)
}

// Post is a utility function which marshals and issues an HTTP post on a specified URL.
func (mock *MockClient) Post(url string, body []byte, output interface{}, headers ...HeaderValue) ([]byte, error) {
	mockArgs := mock.Called(url, body, output, headers)
	return mockArgs.Get(0).([]byte), mockArgs.Error(1)
}
