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
*
* This file was originally developed by The CMux Authors and released under Apache
* License, Version 2.0 in 2016.
************************************************************************************/

package listener

import (
	"strings"
	"testing"
)

func testPTree(t *testing.T, strs ...string) {
	pt := newPatriciaTreeString(strs...)
	for _, s := range strs {
		if !pt.match(strings.NewReader(s)) {
			t.Errorf("%s is not matched by %s", s, s)
		}

		if !pt.matchPrefix(strings.NewReader(s + s)) {
			t.Errorf("%s is not matched as a prefix by %s", s+s, s)
		}

		if pt.match(strings.NewReader(s + s)) {
			t.Errorf("%s matches %s", s+s, s)
		}

		// The following tests are just to catch index out of
		// range and off-by-one errors and not the functionality.
		pt.matchPrefix(strings.NewReader(s[:len(s)-1]))
		pt.match(strings.NewReader(s[:len(s)-1]))
		pt.matchPrefix(strings.NewReader(s + "$"))
		pt.match(strings.NewReader(s + "$"))
	}
}

func TestPatriciaOnePrefix(t *testing.T) {
	testPTree(t, "prefix")
}

func TestPatriciaNonOverlapping(t *testing.T) {
	testPTree(t, "foo", "bar", "dummy")
}

func TestPatriciaOverlapping(t *testing.T) {
	testPTree(t, "foo", "far", "farther", "boo", "ba", "bar")
}
