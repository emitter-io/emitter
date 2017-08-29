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
	"io/ioutil"
	"log"
	"os"
)

// Discard is the discard logger.
var Discard = log.New(ioutil.Discard, "", 0)

// Logger is the logger we use.
var Logger = log.New(os.Stderr, "", log.LstdFlags)

// LogError logs the error as a string.
func LogError(context string, action string, err error) {
	Logger.Printf("[%s] error during %s (%s)\n", context, action, err.Error())
}

// LogAction logs The action with a tag.
func LogAction(context string, action string) {
	Logger.Printf("[%s] %s\n", context, action)
}

// LogTarget logs The action with a tag.
func LogTarget(context, action string, target interface{}) {
	Logger.Printf("[%s] %s (%v)\n", context, action, target)
}
