/**********************************************************************************
* Copyright (c) 2009-2018 Misakai Ltd.
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

package monitor

import (
	"time"
)

// Noop represents a no-op monitor
type Noop struct{}

// NewNoop creates a new no-op monitor.
func NewNoop() *Noop {
	return new(Noop)
}

// Assert contract compliance
var _ Measurer = NewNoop()

// Measure records a value in the queue
func (m *Noop) Measure(name string, value int64) {}

// MeasureElapsed measures elapsed time since the start
func (m *Noop) MeasureElapsed(name string, start time.Time) {}

// MeasureRuntime measures the runtime information
func (m *Noop) MeasureRuntime() {}

// Tag updates a tag.
func (m *Noop) Tag(name, tag string) {}

// Snapshot creates a snapshot
func (m *Noop) Snapshot() []byte { return []byte{} }
