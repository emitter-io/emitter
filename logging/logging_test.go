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
	"encoding/binary"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/quick"
)

var testLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString() string {
	n := rand.Intn(10) + 10

	var ret []rune

	for i := 0; i < n; i++ {
		ret = append(ret, testLetters[rand.Intn(len(testLetters))])
	}

	return string(ret)
}

func TestSetWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	lw := NewWriter()
	lw.SetWriter(buf)

	// simulate some logging
	lw.(*logWriter).w.WriteByte(35)

	lw.SetWriter(&bytes.Buffer{})

	if buf.Bytes()[0] != 35 {
		t.Fatalf("Expected data to be written to the underlying")
	}
}

func TestFlush(t *testing.T) {
	// test that the old one is flushed
	// test new one can be written to
	// probably set twice and check the middle contents
	buf := &bytes.Buffer{}
	lw := NewWriter()
	lw.SetWriter(buf)

	// simulate some logging
	lw.(*logWriter).w.WriteByte(35)

	lw.Flush()

	if buf.Bytes()[0] != 35 {
		t.Fatalf("Expected data to be written to the underlying")
	}
}

func TestAddLogger(t *testing.T) {
	genTest := func(logLine string, expectedKinds []reflect.Kind, expectedSegs []string) func(*testing.T) {
		return func(t *testing.T) {
			if r := recover(); r != nil {
				t.Fatalf("Panic!: %v", r)
			} else {
				t.Logf("No panic.")
			}

			t.Logf("Expected kinds: %v", expectedKinds)
			t.Logf("Expected segs: %v", expectedSegs)

			buf := &bytes.Buffer{}
			lw := NewWriter()
			lw.SetWriter(buf)
			h := lw.AddLogger(logLine)

			//t.Log("Handle:", h)

			lw.Flush()
			out := buf.Bytes()

			//t.Log(string(out))

			expLen := 1 + 4 + 4 + len(expectedKinds)
			for _, s := range expectedSegs {
				expLen += 4 + len(s)
			}

			if len(out) != expLen {
				t.Fatalf("Expected serialized length of %v but got %v.\nOutput: % X", expLen, len(out), out)
			}

			if out[0] != byte(ETLogLine) {
				t.Fatalf("Expected first byte to be ETLogLine but got %v", out[0])
			}

			out = out[1:]

			idbuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(idbuf, uint32(h))

			if !bytes.HasPrefix(out, idbuf) {
				t.Fatalf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
			}

			out = out[4:]

			numSegs := binary.LittleEndian.Uint32(out)

			if numSegs != uint32(len(expectedSegs)) {
				t.Fatalf("Expected %v segments but got %v", len(expectedSegs), numSegs)
			}

			out = out[4:]

			// first check the kinds match
			for i := range expectedKinds {
				k := reflect.Kind(out[0])

				if k != expectedKinds[i] {
					t.Fatalf("Expected kind of %v but got %v", expectedKinds, k)
				}

				out = out[1:]
			}

			for i := range expectedSegs {
				exp := expectedSegs[i]

				seglen := binary.LittleEndian.Uint32(out)

				if seglen != uint32(len(exp)) {
					t.Fatalf("Expected segment length of %v but got %v", len(exp), seglen)
				}

				out = out[4:]

				seg := string(out[:seglen])

				if exp != seg {
					t.Fatalf("Expected segment %v but got %v", exp, seg)
				}

				out = out[seglen:]
			}
		}
	}

	empties := []string{"", ""}

	type testdat struct {
		line     string
		expKinds []reflect.Kind
		expSegs  []string
	}
	type testmap map[string]testdat

	tests := testmap{
		"empty": {
			line:     "",
			expKinds: nil,
			expSegs:  []string{""},
		},
	}

	addKind := func(tm testmap, name, symbol string, kind reflect.Kind) {
		tm[name] = testdat{
			line:     "%" + symbol,
			expKinds: []reflect.Kind{kind},
			expSegs:  empties,
		}
		tm[name+"Brackets"] = testdat{
			line:     "%{" + symbol + "}",
			expKinds: []reflect.Kind{kind},
			expSegs:  empties,
		}

		s1, s2 := randString(), randString()

		tm[name+"WithStrings"] = testdat{
			line:     s1 + "%" + symbol + s2,
			expKinds: []reflect.Kind{kind},
			expSegs:  []string{s1, s2},
		}

		s1, s2 = randString(), randString()

		tm[name+"BracketsWithStrings"] = testdat{
			line:     s1 + "%{" + symbol + "}" + s2,
			expKinds: []reflect.Kind{kind},
			expSegs:  []string{s1, s2},
		}
	}

	addKind(tests, "bool", "b", reflect.Bool)
	addKind(tests, "string", "s", reflect.String)
	addKind(tests, "int", "i", reflect.Int)
	addKind(tests, "int8", "i8", reflect.Int8)
	addKind(tests, "int16", "i16", reflect.Int16)
	addKind(tests, "int32", "i32", reflect.Int32)
	addKind(tests, "int64", "i64", reflect.Int64)
	addKind(tests, "uint", "u", reflect.Uint)
	addKind(tests, "uint8", "u8", reflect.Uint8)
	addKind(tests, "uint16", "u16", reflect.Uint16)
	addKind(tests, "uint32", "u32", reflect.Uint32)
	addKind(tests, "uint64", "u64", reflect.Uint64)
	addKind(tests, "float32", "f32", reflect.Float32)
	addKind(tests, "float64", "f64", reflect.Float64)
	addKind(tests, "complex64", "c64", reflect.Complex64)
	addKind(tests, "complex128", "c128", reflect.Complex128)

	for name, dat := range tests {
		t.Run(name, genTest(dat.line, dat.expKinds, dat.expSegs))
	}
}

func TestAddLoggerLimit(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Correctly got a panic: %v", r)
		} else {
			t.Fatal("Expected a panic but did not get one")
		}
	}()

	lw := NewWriter()
	t.Log("Filling up loggers")
	for i := 0; i < MaxLoggers+1; i++ {
		lw.AddLogger("")
	}
}

func TestParseLogLine(t *testing.T) {
	t.Run("Correct", func(t *testing.T) {
		f := "foo thing bar thing %i64. Fubar %s foo. sadf %% asdf %u32 sdfasfasdfasdffds %u32."
		l, segs := parseLogLine(f)

		// verify logger kinds
		if len(l.Kinds) != 4 {
			t.Fatalf("Expected 4 kinds in logger but got %v", len(l.Kinds))
		}

		// verify logger segs
		if len(segs) != 5 {
			t.Fatalf("Expected 5 segs but got %v", len(segs))
		}
	})

	t.Run("IncorrectFormat", func(t *testing.T) {
		check := func(t *testing.T, fmt string) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Correctly got a panic: %v", r)
				} else {
					t.Fatalf("Expected a panic but did not get one")
				}
			}()

			parseLogLine(fmt)
		}

		tests := map[string]string{}

		addTests := func(name, fmtbase string) {
			tests[name] = "%" + fmtbase
			tests[name+"NotAtEnd"] = "%" + fmtbase + "X"
			tests[name+"MissingEndBrace"] = "%{" + fmtbase
			tests[name+"MissingEndBraceNotAtEnd"] = "%{" + fmtbase + "X"
		}

		addTests("BadInt1", "i1")
		addTests("BadInt3", "i3")
		addTests("BadInt6", "i6")
		addTests("BadUint1", "u1")
		addTests("BadUint3", "u3")
		addTests("BadUint6", "u6")
		addTests("BadFloat", "f")
		addTests("BadFloat32", "f3")
		addTests("BadFloat64", "f6")
		addTests("BadComplex", "c")
		addTests("BadComplex64", "c6")
		addTests("BadComplex128", "c1")
		addTests("BadComplex128Alt", "c12")
		addTests("NoFormatChar", "")

		tests["CorrectButMissingEndBrace"] = "%{b"
		tests["CorrectButMissingEndBraceNotAtEnd"] = "%{bX"

		for name, fmt := range tests {
			t.Run(name, func(t *testing.T) { check(t, fmt) })
		}
	})
}

func TestLog(t *testing.T) {
	check := func(t *testing.T, fmtstring string, toWrite interface{}, dataLen int, checkRest func(*testing.T, []byte) bool) bool {
		// Reset to avoid running over the loggers limit
		buf := &bytes.Buffer{}
		lw := NewWriter()
		lw.SetWriter(buf)
		h := lw.AddLogger(fmtstring)
		//t.Log("Handle:", h)
		lw.Flush()
		buf.Reset()

		lw.Log(h, toWrite)

		lw.Flush()
		out := buf.Bytes()

		expectedLen := 1 + 4 + dataLen

		if len(out) != expectedLen {
			t.Errorf("Expected serialized length of %v but got %v.\nOutput: % X", expectedLen, len(out), out)
			return false
		}

		if out[0] != byte(ETLogEntry) {
			t.Errorf("Expected first byte to be ETLogEntry but got %v", out[0])
			return false
		}

		out = out[1:]

		idbuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(idbuf, uint32(h))

		if !bytes.HasPrefix(out, idbuf) {
			t.Errorf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
			return false
		}

		out = out[4:]

		return checkRest(t, out)
	}

	t.Run("bool", func(t *testing.T) {
		f := func(b bool) bool {
			return check(t, "%b", b, 1, func(t *testing.T, out []byte) bool {
				var expected byte
				if b {
					expected = 1
				}

				if out[0] != expected {
					t.Errorf("Expected %v boolean value to be %v but got %v", b, expected, out[0])
					return false
				}

				return true
			})
		}

		if err := quick.Check(f, nil); err != nil {
			t.Fatalf("Got error during quick test: %v", err)
		}
	})

	t.Run("string", func(t *testing.T) {
		f := func(s string) bool {
			return check(t, "%s", s, 4+len(s), func(t *testing.T, out []byte) bool {
				lenbuf := make([]byte, 4)
				binary.LittleEndian.PutUint32(lenbuf, uint32(len(s)))

				if !bytes.HasPrefix(out, lenbuf) {
					t.Errorf("Expected prefix to match the length of the input string.\nExpected: % X\nGot: % X", lenbuf, out[:4])
					return false
				}

				out = out[4:]

				if string(out) != s {
					t.Errorf("Expected input %v to return but got %v", s, string(out))
					return false
				}

				return true
			})
		}

		if err := quick.Check(f, nil); err != nil {
			t.Fatalf("Got error during quick test: %v", err)
		}
	})

	t.Run("Ints", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			f := func(i int) bool {
				return check(t, "%i", i, 8, func(t *testing.T, out []byte) bool {
					if len(out) != 8 {
						t.Errorf("Expected remaining length to be 8 but got %v", len(out))
					}

					v := int(binary.LittleEndian.Uint64(out))

					if v != i {
						t.Errorf("Expected input %v to return but got %v", i, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("int8", func(t *testing.T) {
			f := func(i int8) bool {
				return check(t, "%i8", i, 1, func(t *testing.T, out []byte) bool {
					if len(out) != 1 {
						t.Errorf("Expected remaining length to be 1 but got %v", len(out))
					}

					v := int8(out[0])

					if v != i {
						t.Errorf("Expected input %v to return but got %v", i, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("int16", func(t *testing.T) {
			f := func(i int16) bool {
				return check(t, "%i16", i, 2, func(t *testing.T, out []byte) bool {
					if len(out) != 2 {
						t.Errorf("Expected remaining length to be 2 but got %v", len(out))
					}

					v := int16(binary.LittleEndian.Uint16(out))

					if v != i {
						t.Errorf("Expected input %v to return but got %v", i, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("int32", func(t *testing.T) {
			f := func(i int32) bool {
				return check(t, "%i32", i, 4, func(t *testing.T, out []byte) bool {
					if len(out) != 4 {
						t.Errorf("Expected remaining length to be 4 but got %v", len(out))
					}

					v := int32(binary.LittleEndian.Uint32(out))

					if v != i {
						t.Errorf("Expected input %v to return but got %v", i, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("int64", func(t *testing.T) {
			f := func(i int64) bool {
				return check(t, "%i64", i, 8, func(t *testing.T, out []byte) bool {
					if len(out) != 8 {
						t.Errorf("Expected remaining length to be 8 but got %v", len(out))
					}

					v := int64(binary.LittleEndian.Uint64(out))

					if v != i {
						t.Errorf("Expected input %v to return but got %v", i, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})
	})

	t.Run("Uints", func(t *testing.T) {
		t.Run("uint", func(t *testing.T) {
			f := func(u uint) bool {
				return check(t, "%u", u, 8, func(t *testing.T, out []byte) bool {
					if len(out) != 8 {
						t.Errorf("Expected remaining length to be 8 but got %v", len(out))
					}

					v := uint(binary.LittleEndian.Uint64(out))

					if v != u {
						t.Errorf("Expected input %v to return but got %v", u, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("uint8", func(t *testing.T) {
			f := func(u uint8) bool {
				return check(t, "%u8", u, 1, func(t *testing.T, out []byte) bool {
					if len(out) != 1 {
						t.Errorf("Expected remaining length to be 1 but got %v", len(out))
					}

					v := uint8(out[0])

					if v != u {
						t.Errorf("Expected input %v to return but got %v", u, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("uint16", func(t *testing.T) {
			f := func(u uint16) bool {
				return check(t, "%u16", u, 2, func(t *testing.T, out []byte) bool {
					if len(out) != 2 {
						t.Errorf("Expected remaining length to be 2 but got %v", len(out))
					}

					v := binary.LittleEndian.Uint16(out)

					if v != u {
						t.Errorf("Expected input %v to return but got %v", u, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("uint32", func(t *testing.T) {
			f := func(u uint32) bool {
				return check(t, "%u32", u, 4, func(t *testing.T, out []byte) bool {
					if len(out) != 4 {
						t.Errorf("Expected remaining length to be 4 but got %v", len(out))
					}

					v := binary.LittleEndian.Uint32(out)

					if v != u {
						t.Errorf("Expected input %v to return but got %v", u, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("uint64", func(t *testing.T) {
			f := func(u uint64) bool {
				return check(t, "%u64", u, 8, func(t *testing.T, out []byte) bool {
					if len(out) != 8 {
						t.Errorf("Expected remaining length to be 8 but got %v", len(out))
					}

					v := binary.LittleEndian.Uint64(out)

					if v != u {
						t.Errorf("Expected input %v to return but got %v", u, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})
	})

	t.Run("Floats", func(t *testing.T) {
		t.Run("float32", func(t *testing.T) {
			f := func(f float32) bool {
				return check(t, "%f32", f, 4, func(t *testing.T, out []byte) bool {
					if len(out) != 4 {
						t.Errorf("Expected remaining length to be 4 but got %v", len(out))
					}

					v := math.Float32frombits(binary.LittleEndian.Uint32(out))

					if v != f {
						t.Errorf("Expected input %v to return but got %v", f, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("float64", func(t *testing.T) {
			f := func(f float64) bool {
				return check(t, "%f64", f, 8, func(t *testing.T, out []byte) bool {
					if len(out) != 8 {
						t.Errorf("Expected remaining length to be 8 but got %v", len(out))
					}

					v := math.Float64frombits(binary.LittleEndian.Uint64(out))

					if v != f {
						t.Errorf("Expected input %v to return but got %v", f, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})
	})

	t.Run("Complexes", func(t *testing.T) {
		t.Run("complex64", func(t *testing.T) {
			f := func(c complex64) bool {
				return check(t, "%c64", c, 8, func(t *testing.T, out []byte) bool {
					if len(out) != 8 {
						t.Errorf("Expected remaining length to be 8 but got %v", len(out))
					}

					v := math.Float32frombits(binary.LittleEndian.Uint32(out))

					if v != real(c) {
						t.Errorf("Expected input %v to return but got %v", c, out)
						return false
					}

					out = out[4:]

					v = math.Float32frombits(binary.LittleEndian.Uint32(out))

					if v != imag(c) {
						t.Errorf("Expected input %v to return but got %v", c, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})

		t.Run("complex128", func(t *testing.T) {
			f := func(c complex128) bool {
				return check(t, "%c128", c, 16, func(t *testing.T, out []byte) bool {
					if len(out) != 16 {
						t.Errorf("Expected remaining length to be 16 but got %v", len(out))
					}

					v := math.Float64frombits(binary.LittleEndian.Uint64(out))

					if v != real(c) {
						t.Errorf("Expected input %v to return but got %v", c, out)
						return false
					}

					out = out[8:]

					v = math.Float64frombits(binary.LittleEndian.Uint64(out))

					if v != imag(c) {
						t.Errorf("Expected input %v to return but got %v", c, out)
						return false
					}

					return true
				})
			}

			if err := quick.Check(f, nil); err != nil {
				t.Fatalf("Got error during quick test: %v", err)
			}
		})
	})

	t.Run("IncorrectUse", func(t *testing.T) {
		t.Run("BadType", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Correctly got a panic: %v", r)
				} else {
					t.Fatalf("Expected a panic but did not get one")
				}
			}()

			lw := NewWriter()
			h := lw.AddLogger("%b")
			//t.Log("Handle:", h)
			lw.Flush()

			lw.Log(h, 42)
		})
		t.Run("WrongNumberOfArgs", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Correctly got a panic: %v", r)
				} else {
					t.Fatalf("Expected a panic but did not get one")
				}
			}()

			lw := NewWriter()
			h := lw.AddLogger("%b")
			//t.Log("Handle:", h)
			lw.Flush()

			lw.Log(h, true, 42)
		})
	})
}

var testLogHandleSink Handle

func BenchmarkAddLogger(b *testing.B) {
	lw := NewWriter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLogHandleSink = lw.AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")

		// to prevent it from overflowing the logger array
		*lw.(*logWriter).curLoggersIdx = 0
	}
}

var (
	testLoggerSink   Logger
	testSegmentsSink []string
)

func BenchmarkParseLogLine(b *testing.B) {
	f := "The operation %s could not be completed. Wanted %u64 bar %c128 %b %{s} %{i32}"
	for i := 0; i < b.N; i++ {
		testLoggerSink, testSegmentsSink = parseLogLine(f)
	}
}

func BenchmarkLogParallel(b *testing.B) {
	lw := NewWriter()
	h := lw.AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")
	args := []interface{}{int64(1), "string", uint32(2), uint32(3)}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lw.Log(h, args...)
		}
	})
}

func BenchmarkLogSequential(b *testing.B) {
	lw := NewWriter()
	h := lw.AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")
	args := []interface{}{int64(1), "string", uint32(2), uint32(3)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lw.Log(h, args...)
	}
}

func BenchmarkCompareToStdlib(b *testing.B) {
	b.Run("Nanolog", func(b *testing.B) {
		lw := NewWriter()
		h := lw.AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")
		args := []interface{}{int64(1), "string", uint32(2), uint32(3)}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			lw.Log(h, args...)
		}
	})
	b.Run("Stdlib", func(b *testing.B) {
		args := []interface{}{int64(1), "string", uint32(2), uint32(3)}
		l := log.New(ioutil.Discard, "", 0)
		for i := 0; i < b.N; i++ {
			l.Printf("foo thing bar thing %d. Fubar %s foo. sadfasdf %d sdfasfasdfasdffds %d.", args...)
		}
	})
}

func BenchmarkInterpolations(b *testing.B) {

	f := func(interp string, item interface{}, limit int) func(b *testing.B) {
		return func(b *testing.B) {
			for i := 1; i <= limit; i++ {
				b.Run(strconv.Itoa(i), func(b *testing.B) {
					lw := NewWriter()
					h := lw.AddLogger(strings.Repeat(interp, i))
					args := make([]interface{}, i)

					for j := range args {
						args[j] = interface{}(item)
					}

					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						Log(h, args...)
					}
				})
			}
		}
	}

	b.Run("Single", func(b *testing.B) {
		b.Run("bool", f("%b", true, 1))
		b.Run("string", f("%s", "this is a string", 1))
		b.Run("int", f("%i", 4, 1))
		b.Run("int8", f("%i8", int8(4), 1))
		b.Run("int16", f("%i16", int16(4), 1))
		b.Run("int32", f("%i32", int32(4), 1))
		b.Run("int64", f("%i64", int64(4), 1))
		b.Run("uint", f("%u", uint(4), 1))
		b.Run("uint8", f("%u8", uint8(4), 1))
		b.Run("uint16", f("%u16", uint16(4), 1))
		b.Run("uint32", f("%u32", uint32(4), 1))
		b.Run("uint64", f("%u64", uint64(4), 1))
		b.Run("float32", f("%f32", float32(4), 1))
		b.Run("float64", f("%f64", float64(4), 1))
		b.Run("complex64", f("%c64", complex(float32(4), float32(4)), 1))
		b.Run("complex128", f("%c128", complex(float64(4), float64(4)), 1))
	})

	b.Run("Multi", func(b *testing.B) {
		b.Run("bool", f("%b", true, 10))
		b.Run("string", f("%s", "this is a string", 10))
		b.Run("int", f("%i", 4, 10))
		b.Run("int8", f("%i8", int8(4), 10))
		b.Run("int16", f("%i16", int16(4), 10))
		b.Run("int32", f("%i32", int32(4), 10))
		b.Run("int64", f("%i64", int64(4), 10))
		b.Run("uint", f("%u", uint(4), 10))
		b.Run("uint8", f("%u8", uint8(4), 10))
		b.Run("uint16", f("%u16", uint16(4), 10))
		b.Run("uint32", f("%u32", uint32(4), 10))
		b.Run("uint64", f("%u64", uint64(4), 10))
		b.Run("float32", f("%f32", float32(4), 10))
		b.Run("float64", f("%f64", float64(4), 10))
		b.Run("complex64", f("%c64", complex(float32(4), float32(4)), 10))
		b.Run("complex128", f("%c128", complex(float64(4), float64(4)), 10))
	})
}
