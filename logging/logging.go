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

// Package logging is a package to speed up your logging.
//
// The format string is inspired by the full fledged fmt.Fprintf function. The
// codes are unique to this package, so normal fmt documentation is not be applicable.
//
// The format string is similar to fmt in that it uses the percent sign (a.k.a.
// the modulo operator) to signify the start of a format code. The reader is
// greedy, meaning that the parser will attempt to read as much as it can for a
// code before it stops. E.g. if you have a generic int in the middle of your
// format string immediately followed by the number 1 and a space ("%i1 "), the
// parser may complain saying that it encountered an invalid code. To fix this,
// use curly braces after the percent sign to surround the code: "%{i}1 ".
//
// Kinds and their corresponding format codes:
//
//  Kind          | Code
//  --------------|-------------
//  Bool          | b
//  Int           | i
//  Int8          | i8
//  Int16         | i16
//  Int32         | i32
//  Int64         | i64
//  Uint          | u
//  Uint8         | u8
//  Uint16        | u16
//  Uint32        | u32
//  Uint64        | u64
//  Uintptr       |
//  Float32       | f32
//  Float64       | f64
//  Complex64     | c64
//  Complex128    | c128
//  Array         |
//  Chan          |
//  Func          |
//  Interface     |
//  Map           |
//  Ptr           |
//  Slice         |
//  String        | s
//  Struct        |
//  UnsafePointer |
//
// The file format has two categories of data:
//
//  1. Log line information to reconstruct logs later
//  2. Actual log entries
//
// The differentiation is done with the entryType, which is prefixed on to the record.
//
// The log line records are formatted as follows:
//
//  - type:             1 byte - ETLogLine (1)
//  - id:               4 bytes - little endian uint32
//  - # of string segs: 4 bytes - little endian uint32
//  - kinds:            (#segs - 1) bytes, each being a reflect.Kind
//  - segments:
//    - string length:  4 bytes - little endian uint32
//    - string data:    ^length bytes
//
// The log entry records are formatted as follows:
//
//  - type:    1 byte - ETLogEntry (2)
//  - line id: 4 bytes - little endian uint32
//  - data+:   var bytes - all the corresponding data for the kinds in the log line entry
//
// The data is serialized as follows:
//
//  - Bool: 1 byte
//    - False: 0 or True: 1
//
//  - String: 4 + len(string) bytes
//    - Length: 4 bytes - little endian uint32
//    - String bytes: Length bytes
//
//  - int family:
//     - int:    8 bytes - int64 as little endian uint64
//     - int8:  1 byte
//     - int16: 2 bytes - int16 as little endian uint16
//     - int32: 4 bytes - int32 as little endian uint32
//     - int64: 8 bytes - int64 as little endian uint64
//
//  - uint family:
//    - uint:   8 bytes - little endian uint64
//    - uint8:  1 byte
//    - uint16: 2 bytes - little endian uint16
//    - uint32: 4 bytes - little endian uint32
//    - uint64: 8 bytes - little endian uint64
//
//  - float32:
//    - 4 bytes as little endian uint32 from float32 bits
//
//  - float64:
//    - 8 bytes as little endian uint64 from float64 bits
//
//  - complex64:
//    - Real:    4 bytes as little endian uint32 from float32 bits
//    - Complex: 4 bytes as little endian uint32 from float32 bits
//
//  - complex128:
//    - Real:    8 bytes as little endian uint64 from float64 bits
//    - Complex: 8 bytes as little endian uint64 from float64 bits
package logging

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"unicode/utf8"
)

// MaxLoggers is the maximum number of different loggers that are allowed
const MaxLoggers = 10240

// Handle is a simple handle to an internal logging data structure
// LogHandles are returned by the AddLogger method and used by the Log method to
// actually log data.
type Handle uint32

// EntryType is an enum that represents the record headers in the output files to
// differentiate between log lines and log entries
type EntryType byte

const (
	// ETInvalid is an invalid EntryType
	ETInvalid EntryType = iota

	// ETLogLine means the log line data for a single call to AddLogger is ahead
	ETLogLine

	// ETLogEntry means the log data for a single call to Log is ahead
	ETLogEntry
)

// Logger is the internal struct representing the runtime state of the loggers.
// The Segs field is not used during logging; it is only used in the inflate
// utility
type Logger struct {
	Kinds []reflect.Kind
	Segs  []string
}

var defaultLogWriter = NewWriter()

// LogWriter represents an interface which logger implements.
type LogWriter interface {
	// SetWriter will set up efficient writing for the log to the output stream given.
	// A raw IO stream is best. The first time SetWriter is called any logs that were
	// created or posted before the call will be sent to the writer all in one go.
	SetWriter(new io.Writer) error

	// Flush ensures all log entries written up to this point are written to the underlying io.Writer
	Flush() error

	// AddLogger initializes a logger and returns a handle for future logging
	AddLogger(fmt string) Handle

	// Log logs to the output stream
	Log(handle Handle, args ...interface{}) error
}

type logWriter struct {
	initBuf       *bytes.Buffer
	w             *bufio.Writer
	firstSet      bool
	writeLock     sync.Locker
	loggers       []Logger
	curLoggersIdx *uint32
}

// Error represents an error logging handler
var (
	logError  = AddLogger("[%s] error during %s (%s)")
	logAction = AddLogger("[%s] %s")
)

// LogError logs the error as a string.
func LogError(context string, action string, err error) {
	Log(logError, context, action, err.Error())
}

// LogAction logs The action with a tag.
func LogAction(context string, action string) {
	Log(logAction, context, action)
}

// NewWriter creates a new LogWriter
func NewWriter() LogWriter {
	initBuf := &bytes.Buffer{}
	return &logWriter{
		initBuf:       initBuf,
		w:             bufio.NewWriter(initBuf),
		firstSet:      true,
		writeLock:     new(sync.Mutex),
		loggers:       make([]Logger, MaxLoggers),
		curLoggersIdx: new(uint32),
	}
}

// SetWriter calls LogWriter.SetWriter on the default log writer.
func SetWriter(new io.Writer, inflated bool) error {
	if inflated {
		return defaultLogWriter.SetWriter(InflateTo(new))
	}

	return defaultLogWriter.SetWriter(new)
}

func (lw *logWriter) SetWriter(new io.Writer) error {
	// grab write lock to ensure no problems
	lw.writeLock.Lock()
	defer lw.writeLock.Unlock()

	if err := lw.w.Flush(); err != nil {
		return err
	}

	lw.w = bufio.NewWriter(new)

	if lw.firstSet {
		lw.firstSet = false
		if _, err := lw.initBuf.WriteTo(lw.w); err != nil {
			return err
		}
	}

	return nil
}

// Flush calls LogWriter.Flush on the default log writer.
func Flush() error {
	return defaultLogWriter.Flush()
}

func (lw *logWriter) Flush() error {
	// grab write lock to ensure no prblems
	lw.writeLock.Lock()
	defer lw.writeLock.Unlock()

	return lw.w.Flush()
}

// AddLogger calls LogWriter.AddLogger on the default log writer.
func AddLogger(fmt string) Handle {
	return defaultLogWriter.AddLogger(fmt)
}

func (lw *logWriter) AddLogger(fmt string) Handle {
	// save some kind of string format to the file
	idx := atomic.AddUint32(lw.curLoggersIdx, 1) - 1

	if idx >= MaxLoggers {
		panic("Too many loggers")
	}

	l, segs := parseLogLine(fmt)
	lw.loggers[idx] = l

	lw.writeLogDataToFile(idx, l.Kinds, segs)

	return Handle(idx)
}

func parseLogLine(gold string) (Logger, []string) {
	// make a copy we can destroy
	tmp := gold
	f := &tmp
	var kinds []reflect.Kind
	var segs []string
	var curseg []rune

	for len(*f) > 0 {
		if r := next(f); r != '%' {
			curseg = append(curseg, r)
			continue
		}

		// Literal % sign
		if peek(f) == '%' {
			next(f)
			curseg = append(curseg, '%')
			continue
		}

		segs = append(segs, string(curseg))
		curseg = curseg[:0]

		var requireBrace bool

		// Optional curly braces around format
		r := next(f)
		if r == '{' {
			requireBrace = true
			r = next(f)
		}

		// optimized parse tree
		switch r {
		case 'b':
			kinds = append(kinds, reflect.Bool)

		case 's':
			kinds = append(kinds, reflect.String)

		case 'i':
			if len(*f) == 0 {
				kinds = append(kinds, reflect.Int)
				break
			}

			r := peek(f)
			switch r {
			case '8':
				next(f)
				kinds = append(kinds, reflect.Int8)

			case '1':
				next(f)
				if next(f) != '6' {
					logpanic("Was expecting i16.", gold)
				}
				kinds = append(kinds, reflect.Int16)

			case '3':
				next(f)
				if next(f) != '2' {
					logpanic("Was expecting i32.", gold)
				}
				kinds = append(kinds, reflect.Int32)

			case '6':
				next(f)
				if next(f) != '4' {
					logpanic("Was expecting i64.", gold)
				}
				kinds = append(kinds, reflect.Int64)

			default:
				kinds = append(kinds, reflect.Int)
			}

		case 'u':
			if len(*f) == 0 {
				kinds = append(kinds, reflect.Uint)
				break
			}

			r := peek(f)
			switch r {
			case '8':
				next(f)
				kinds = append(kinds, reflect.Uint8)

			case '1':
				next(f)
				if next(f) != '6' {
					logpanic("Was expecting u16.", gold)
				}
				kinds = append(kinds, reflect.Uint16)

			case '3':
				next(f)
				if next(f) != '2' {
					logpanic("Was expecting u32.", gold)
				}
				kinds = append(kinds, reflect.Uint32)

			case '6':
				next(f)
				if next(f) != '4' {
					logpanic("Was expecting u64.", gold)
				}
				kinds = append(kinds, reflect.Uint64)

			default:
				kinds = append(kinds, reflect.Uint)
			}

		case 'f':
			r := peek(f)
			switch r {
			case '3':
				next(f)
				if next(f) != '2' {
					logpanic("Was expecting f32.", gold)
				}
				kinds = append(kinds, reflect.Float32)

			case '6':
				next(f)
				if next(f) != '4' {
					logpanic("Was expecting f64.", gold)
				}
				kinds = append(kinds, reflect.Float64)

			default:
				logpanic("Expecting either f32 or f64", gold)
			}

		case 'c':
			r := peek(f)
			switch r {
			case '6':
				next(f)
				if next(f) != '4' {
					logpanic("Was expecting c64.", gold)
				}
				kinds = append(kinds, reflect.Complex64)

			case '1':
				next(f)
				if next(f) != '2' {
					logpanic("Was expecting c128.", gold)
				}
				if next(f) != '8' {
					logpanic("Was expecting c128.", gold)
				}
				kinds = append(kinds, reflect.Complex128)

			default:
				logpanic("Expecting either c64 or c128", gold)
			}

		default:
			logpanic("Invalid replace sequence", gold)
		}

		if requireBrace {
			if len(*f) == 0 {
				logpanic("Missing '}' character at end of line", gold)
			}
			if next(f) != '}' {
				logpanic("Missing '}' character", gold)
			}
		}
	}

	segs = append(segs, string(curseg))

	return Logger{
		Kinds: kinds,
	}, segs
}

func peek(s *string) rune {
	r, _ := utf8.DecodeRuneInString(*s)

	if r == utf8.RuneError {
		panic("Malformed log string")
	}

	return r
}

func next(s *string) rune {
	r, n := utf8.DecodeRuneInString(*s)
	*s = (*s)[n:]

	if r == utf8.RuneError {
		panic("Malformed log string")
	}

	return r
}

func (lw *logWriter) writeLogDataToFile(idx uint32, kinds []reflect.Kind, segs []string) {
	buf := &bytes.Buffer{}
	b := make([]byte, 4)

	// write log line record identifier
	buf.WriteByte(byte(ETLogLine))

	// write log identifier
	binary.LittleEndian.PutUint32(b, idx)
	buf.Write(b)

	// write number of string segments between variable parts
	// we don't need to write the number of kinds here because it is always
	// equal to the number of segments minus 1
	if len(segs) > math.MaxInt32 {
		// what the hell are you logging?!
		panic("Too many log line segments")
	}
	binary.LittleEndian.PutUint32(b, uint32(len(segs)))
	buf.Write(b)

	// write out all the kinds. These are cast to a byte because their values all
	// fit into a byte and it saves a little space
	for _, k := range kinds {
		buf.WriteByte(byte(k))
	}

	// write all the segments, lengths first then string bytes for each
	for _, s := range segs {
		binary.LittleEndian.PutUint32(b, uint32(len(s)))
		buf.Write(b)
		buf.WriteString(s)
	}

	// finally write all of it together to the output
	lw.w.Write(buf.Bytes())
}

// helper function to have consistently formatted panics and shorter code above
func logpanic(msg, gold string) {
	panic(fmt.Sprintf("Malformed log format string. %s.\n%s", msg, gold))
}

var (
	bufpool = &sync.Pool{
		New: func() interface{} {
			temp := make([]byte, 1024) // 1k default size
			return &temp
		},
	}
)

// Log calls LogWriter.Log on the default log writer.
func Log(handle Handle, args ...interface{}) error {
	return defaultLogWriter.Log(handle, args...)
}

func (lw *logWriter) Log(handle Handle, args ...interface{}) error {
	l := lw.loggers[handle]

	if len(l.Kinds) != len(args) {
		panic("Number of args does not match log line")
	}

	buf := bufpool.Get().(*[]byte)
	*buf = (*buf)[:0]
	b := make([]byte, 8)

	*buf = append(*buf, byte(ETLogEntry))

	binary.LittleEndian.PutUint32(b, uint32(handle))
	*buf = append(*buf, b[:4]...)

	for idx := range l.Kinds {
		if l.Kinds[idx] != reflect.TypeOf(args[idx]).Kind() {
			panic("Argument type does not match log line")
		}

		// write serialized version to writer
		switch l.Kinds[idx] {
		case reflect.Bool:
			if args[idx].(bool) {
				*buf = append(*buf, 1)
			} else {
				*buf = append(*buf, 0)
			}

		case reflect.String:
			s := args[idx].(string)
			binary.LittleEndian.PutUint32(b, uint32(len(s)))
			*buf = append(*buf, b[:4]...)
			*buf = append(*buf, s...)

		// ints
		case reflect.Int:
			// Assume generic int is 64 bit
			i := args[idx].(int)
			binary.LittleEndian.PutUint64(b, uint64(i))
			*buf = append(*buf, b...)

		case reflect.Int8:
			i := args[idx].(int8)
			*buf = append(*buf, byte(i))

		case reflect.Int16:
			i := args[idx].(int16)
			binary.LittleEndian.PutUint16(b, uint16(i))
			*buf = append(*buf, b[:2]...)

		case reflect.Int32:
			i := args[idx].(int32)
			binary.LittleEndian.PutUint32(b, uint32(i))
			*buf = append(*buf, b[:4]...)

		case reflect.Int64:
			i := args[idx].(int64)
			binary.LittleEndian.PutUint64(b, uint64(i))
			*buf = append(*buf, b...)

		// uints
		case reflect.Uint:
			// Assume generic uint is 64 bit
			i := args[idx].(uint)
			binary.LittleEndian.PutUint64(b, uint64(i))
			*buf = append(*buf, b...)

		case reflect.Uint8:
			i := args[idx].(uint8)
			*buf = append(*buf, byte(i))

		case reflect.Uint16:
			i := args[idx].(uint16)
			binary.LittleEndian.PutUint16(b, i)
			*buf = append(*buf, b[:2]...)

		case reflect.Uint32:
			i := args[idx].(uint32)
			binary.LittleEndian.PutUint32(b, i)
			*buf = append(*buf, b[:4]...)

		case reflect.Uint64:
			i := args[idx].(uint64)
			binary.LittleEndian.PutUint64(b, i)
			*buf = append(*buf, b...)

		// floats
		case reflect.Float32:
			f := args[idx].(float32)
			i := math.Float32bits(f)
			binary.LittleEndian.PutUint32(b, i)
			*buf = append(*buf, b[:4]...)

		case reflect.Float64:
			f := args[idx].(float64)
			i := math.Float64bits(f)
			binary.LittleEndian.PutUint64(b, i)
			*buf = append(*buf, b...)

		// complex
		case reflect.Complex64:
			c := args[idx].(complex64)

			f := real(c)
			i := math.Float32bits(f)
			binary.LittleEndian.PutUint32(b, i)
			*buf = append(*buf, b[:4]...)

			f = imag(c)
			i = math.Float32bits(f)
			binary.LittleEndian.PutUint32(b, i)
			*buf = append(*buf, b[:4]...)

		case reflect.Complex128:
			c := args[idx].(complex128)

			f := real(c)
			i := math.Float64bits(f)
			binary.LittleEndian.PutUint64(b, i)
			*buf = append(*buf, b...)

			f = imag(c)
			i = math.Float64bits(f)
			binary.LittleEndian.PutUint64(b, i)
			*buf = append(*buf, b...)

		default:
			panic(fmt.Sprintf("Invalid Kind in logger: %v", l.Kinds[idx]))
		}
	}

	lw.writeLock.Lock()
	_, err := lw.w.Write(*buf)
	lw.writeLock.Unlock()

	bufpool.Put(buf)
	return err
}
