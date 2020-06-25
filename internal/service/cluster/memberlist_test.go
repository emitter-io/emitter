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

package cluster

import (
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/mesh"
)

func TestFallback(t *testing.T) {
	m := newMemberlist(newPeer)
	_, ok := m.Fallback(2000)
	assert.False(t, ok)

	m.GetOrAdd(1900)
	m.GetOrAdd(2000)
	m.GetOrAdd(2500)
	m.GetOrAdd(2200)

	f, ok := m.Fallback(2000)
	assert.True(t, ok)
	assert.Equal(t, 2200, int(f.name))
}

func newPeer(name mesh.PeerName) *Peer {
	return &Peer{
		name:     name,
		frame:    message.NewFrame(defaultFrameSize),
		subs:     message.NewCounters(),
		activity: time.Now().Unix(),
	}
}
