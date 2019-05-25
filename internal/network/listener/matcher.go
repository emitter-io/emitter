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
	"bytes"
	"io"
)

var defaultHTTPMethods = []string{
	"OPTIONS",
	"GET",
	"HEAD",
	"POST",
	"PATCH",
	"PUT",
	"DELETE",
	"TRACE",
	"CONNECT",
}

// ------------------------------------------------------------------------------------

// MatchAny matches any connection.
func MatchAny() Matcher {
	return func(r io.Reader) bool { return true }
}

// MatchPrefix returns a matcher that matches a connection if it
// starts with any of the strings in strs.
func MatchPrefix(strs ...string) Matcher {
	pt := newPatriciaTreeString(strs...)
	return pt.matchPrefix
}

// MatchHTTP only matches the methods in the HTTP request.
func MatchHTTP(extMethods ...string) Matcher {
	return MatchPrefix(append(defaultHTTPMethods, extMethods...)...)
}

// ------------------------------------------------------------------------------------

// patriciaTree is a simple patricia tree that handles []byte instead of string
// and cannot be changed after instantiation.
type patriciaTree struct {
	root     *ptNode
	maxDepth int // max depth of the tree.
}

func newPatriciaTree(bs ...[]byte) *patriciaTree {
	max := 0
	for _, b := range bs {
		if max < len(b) {
			max = len(b)
		}
	}
	return &patriciaTree{
		root:     newNode(bs),
		maxDepth: max + 1,
	}
}

func newPatriciaTreeString(strs ...string) *patriciaTree {
	b := make([][]byte, len(strs))
	for i, s := range strs {
		b[i] = []byte(s)
	}
	return newPatriciaTree(b...)
}

func (t *patriciaTree) matchPrefix(r io.Reader) bool {
	buf := make([]byte, t.maxDepth)
	n, _ := io.ReadFull(r, buf)
	return t.root.match(buf[:n], true)
}

func (t *patriciaTree) match(r io.Reader) bool {
	buf := make([]byte, t.maxDepth)
	n, _ := io.ReadFull(r, buf)
	return t.root.match(buf[:n], false)
}

type ptNode struct {
	prefix   []byte
	next     map[byte]*ptNode
	terminal bool
}

func newNode(strs [][]byte) *ptNode {
	if len(strs) == 0 {
		return &ptNode{
			prefix:   []byte{},
			terminal: true,
		}
	}

	if len(strs) == 1 {
		return &ptNode{
			prefix:   strs[0],
			terminal: true,
		}
	}

	p, strs := splitPrefix(strs)
	n := &ptNode{
		prefix: p,
	}

	nexts := make(map[byte][][]byte)
	for _, s := range strs {
		if len(s) == 0 {
			n.terminal = true
			continue
		}
		nexts[s[0]] = append(nexts[s[0]], s[1:])
	}

	n.next = make(map[byte]*ptNode)
	for first, rests := range nexts {
		n.next[first] = newNode(rests)
	}

	return n
}

func splitPrefix(bss [][]byte) (prefix []byte, rest [][]byte) {
	if len(bss) == 0 || len(bss[0]) == 0 {
		return prefix, bss
	}

	if len(bss) == 1 {
		return bss[0], [][]byte{{}}
	}

	for i := 0; ; i++ {
		var cur byte
		eq := true
		for j, b := range bss {
			if len(b) <= i {
				eq = false
				break
			}

			if j == 0 {
				cur = b[i]
				continue
			}

			if cur != b[i] {
				eq = false
				break
			}
		}

		if !eq {
			break
		}

		prefix = append(prefix, cur)
	}

	rest = make([][]byte, 0, len(bss))
	for _, b := range bss {
		rest = append(rest, b[len(prefix):])
	}

	return prefix, rest
}

func (n *ptNode) match(b []byte, prefix bool) bool {
	l := len(n.prefix)
	if l > 0 {
		if l > len(b) {
			l = len(b)
		}
		if !bytes.Equal(b[:l], n.prefix) {
			return false
		}
	}

	if n.terminal && (prefix || len(n.prefix) == len(b)) {
		return true
	}

	if l >= len(b) {
		return false
	}

	nextN, ok := n.next[b[l]]
	if !ok {
		return false
	}

	if l == len(b) {
		b = b[l:l]
	} else {
		b = b[l+1:]
	}
	return nextN.match(b, prefix)
}
