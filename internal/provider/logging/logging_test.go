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

package logging

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestLogger creates a new default stderr logger.
func newTestLogger(buffer *bytes.Buffer) Logging {
	return (*stderrLogger)(log.New(buffer, "", 0))
}

func TestLogAction(t *testing.T) {
	defer func(l Logging) { Logger = l }(Logger)

	buffer := bytes.NewBuffer(nil)
	Logger = newTestLogger(buffer)

	LogAction("a", "b")
	assert.Equal(t, "[a] b\n", string(buffer.Bytes()))
}

func TestLogError(t *testing.T) {
	defer func(l Logging) { Logger = l }(Logger)

	buffer := bytes.NewBuffer(nil)
	Logger = newTestLogger(buffer)

	LogError("a", "b", errors.New("err"))
	assert.Equal(t, "[a] error during b (err)\n", string(buffer.Bytes()))
}

func TestLogTarget(t *testing.T) {
	defer func(l Logging) { Logger = l }(Logger)

	buffer := bytes.NewBuffer(nil)
	Logger = newTestLogger(buffer)

	LogTarget("a", "b", 123)
	assert.Equal(t, "[a] b (123)\n", string(buffer.Bytes()))
}

func TestStdErrLogger(t *testing.T) {
	l := NewStdErr()
	assert.NoError(t, l.Configure(nil))
	assert.Equal(t, "stderr", l.Name())
}
