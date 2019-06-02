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
	"fmt"

	"github.com/emitter-io/emitter/internal/provider/logging"
	"github.com/emitter-io/emitter/internal/security/license"
	"github.com/jawher/mow.cli"
)

// New generates a new license and secret key pair.
func New(cmd *cli.Cmd) {
	cmd.Spec = ""
	cmd.Action = func() {
		license, secret := license.New()
		logging.LogAction("license", fmt.Sprintf("generated new license: %v", license))
		logging.LogAction("license", fmt.Sprintf("generated new secret key: %v", secret))
	}
}
