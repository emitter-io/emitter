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

package version

import (
	"fmt"

	"github.com/emitter-io/emitter/internal/provider/logging"
	cli "github.com/jawher/mow.cli"
)

var (
	version = "0"
	commit  = "untracked"
)

// Print prints the version
func Print(cmd *cli.Cmd) {
	cmd.Spec = ""
	cmd.Action = func() {
		logging.LogAction("version", fmt.Sprintf("emitter version %s, commit %s", version, commit))
	}
}
