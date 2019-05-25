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
	"io/ioutil"
	"log"
	"os"

	"github.com/emitter-io/config"
)

// Discard is the discard logger.
var Discard = log.New(ioutil.Discard, "", 0)

// Logger is the logger we use.
var Logger = NewStdErr()

// LogError logs the error as a string.
func LogError(context string, action string, err error) {
	Logger.Printf("[%s] error during %s (%s)", context, action, err.Error())
}

// LogAction logs The action with a tag.
func LogAction(context string, action string) {
	Logger.Printf("[%s] %s", context, action)
}

// LogTarget logs The action with a tag.
func LogTarget(context, action string, target interface{}) {
	Logger.Printf("[%s] %s (%v)", context, action, target)
}

// ------------------------------------------------------------------------------------

// Logging represents a logging contract.
type Logging interface {
	config.Provider

	// Printf logs a formatted log line.
	Printf(format string, v ...interface{})
}

// ------------------------------------------------------------------------------------

// stderrLogger implements Logging contract.
var _ Logging = new(stderrLogger)

// stderrLogger represents a simple golang logger.
type stderrLogger log.Logger

// NewStdErr creates a new default stderr logger.
func NewStdErr() Logging {
	return (*stderrLogger)(log.New(os.Stderr, "", log.LstdFlags))
}

// Name returns the name of the provider.
func (s *stderrLogger) Name() string {
	return "stderr"
}

// Configure configures the provider
func (s *stderrLogger) Configure(config map[string]interface{}) error {
	return nil
}

// Printf prints a log line.
func (s *stderrLogger) Printf(format string, v ...interface{}) {
	(*log.Logger)(s).Printf(format+"\n", v...)
}
