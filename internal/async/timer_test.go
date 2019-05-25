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

package async

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRepeat(t *testing.T) {
	assert.NotPanics(t, func() {
		out := make(chan bool, 1)
		cancel := Repeat(context.TODO(), time.Nanosecond*10, func() {
			out <- true
		})

		<-out
		v := <-out
		assert.True(t, v)
		cancel()
	})
}

func TestRepeatFirstActionPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		cancel := Repeat(context.TODO(), time.Nanosecond*10, func() {
			panic("test")
		})

		cancel()
	})
}

func TestRepeatPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		var counter int32

		cancel := Repeat(context.TODO(), time.Nanosecond*10, func() {
			atomic.AddInt32(&counter, 1)
			panic("test")
		})

		for atomic.LoadInt32(&counter) <= 10 {
		}

		cancel()
	})
}
