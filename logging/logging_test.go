/**********************************************************************************
* Copyright (c) 2009-2017 Misakai Ltd.
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

package logging

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogAction(t *testing.T) {
	defer func(l *log.Logger) { logger = l }(logger)

	buffer := bytes.NewBuffer(nil)
	logger = log.New(buffer, "", 0)

	LogAction("a", "b")
	assert.Equal(t, "[a] b\n", string(buffer.Bytes()))
}

func TestLogError(t *testing.T) {
	defer func(l *log.Logger) { logger = l }(logger)

	buffer := bytes.NewBuffer(nil)
	logger = log.New(buffer, "", 0)

	LogError("a", "b", errors.New("err"))
	assert.Equal(t, "[a] error during b (err)\n", string(buffer.Bytes()))
}

func TestLogTarget(t *testing.T) {
	defer func(l *log.Logger) { logger = l }(logger)

	buffer := bytes.NewBuffer(nil)
	logger = log.New(buffer, "", 0)

	LogTarget("a", "b", 123)
	assert.Equal(t, "[a] b (123)\n", string(buffer.Bytes()))
}
