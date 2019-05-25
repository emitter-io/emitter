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
	"fmt"
	"runtime/debug"
	"time"

	"github.com/emitter-io/emitter/internal/provider/logging"
)

// Repeat performs an action asynchronously on a predetermined interval.
func Repeat(ctx context.Context, interval time.Duration, action func()) context.CancelFunc {

	// Create cancellation context first
	ctx, cancel := context.WithCancel(ctx)
	safeAction := func() {
		defer handlePanic()
		action()
	}

	// Perform the action for the first time, syncrhonously
	safeAction()
	timer := time.NewTicker(interval)
	go func() {

		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				safeAction()
			}
		}
	}()

	return cancel
}

// handlePanic handles the panic and logs it out.
func handlePanic() {
	if r := recover(); r != nil {
		logging.LogAction("async", fmt.Sprintf("panic recovered: %ss \n %s", r, debug.Stack()))
	}
}
