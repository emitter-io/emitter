// Copyright 2017 Scott Mansfield
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	inbuf := &bytes.Buffer{}
	SetWriter(inbuf, false)

	// This should exercise every different type in one log line
	h := AddLogger("%b %s %i %i8 %i16 %i32 %i64 %u %u8 %u16 %u32 %u64 %f32 %f64 %c64 %c128")
	Log(h,
		true, "",
		int(4), int8(4), int16(4), int32(4), int64(4),
		uint(4), uint8(4), uint16(4), uint32(4), uint64(4),
		float32(4), float64(4),
		complex(float32(4), float32(4)), complex(float64(4), float64(4)),
	)
	Flush()

	outbuf := &bytes.Buffer{}
	r := NewReader(outbuf)
	if err := r.Inflate(inbuf.Bytes()); err != nil {
		t.Fatalf("Got error during inflate: %v", err)
	}
}

func TestReaderEarlyExits(t *testing.T) {
	inbuf := &bytes.Buffer{}
	SetWriter(inbuf, false)

	// This should exercise every different type in one log line
	h := AddLogger("%b %s %i %i8 %i16 %i32 %i64 %u %u8 %u16 %u32 %u64 %f32 %f64 %c64 %c128")
	Log(h,
		false, "",
		int(4), int8(4), int16(4), int32(4), int64(4),
		uint(4), uint8(4), uint16(4), uint32(4), uint64(4),
		float32(4), float64(4),
		complex(float32(4), float32(4)), complex(float64(4), float64(4)),
	)
	Flush()

	outbuf := &bytes.Buffer{}

	for i := 0; i < inbuf.Len(); i++ {
		t.Run(fmt.Sprintf("Exit%d", i), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic!: %v", r)
				} else {
					t.Logf("No panic.")
				}
			}()

			outbuf.Reset()
			r := NewReader(outbuf)
			if err := r.Inflate(inbuf.Bytes()); err != nil {
				t.Logf("Got error during inflate: %v", err)
			}
		})
	}
}

func TestInflateTo(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = InflateTo(os.Stdout)
	})
}
