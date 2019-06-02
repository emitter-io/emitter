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

package license

import (
	"testing"

	"github.com/jawher/mow.cli"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() {
		runCommand(New)
	})
}

func runCommand(f func(cmd *cli.Cmd), args ...string) {
	app := cli.App("emitter", "")
	app.Command("test", "", f)
	v := []string{"emitter", "test"}
	v = append(v, args...)
	app.Run(v)
}
