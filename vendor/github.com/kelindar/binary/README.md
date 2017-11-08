# Generic and Fast Binary Serializer for Go

This repository contains a fast binary packer for Golang, this allows to encode/decode arbtitrary golang data structures of variable size. [Documentation](https://godoc.org/github.com/Kelindar/binary) can be found on [https://godoc.org/github.com/Kelindar/binary](https://godoc.org/github.com/Kelindar/binary).

This package extends support to arbitrary, variable-sized values by prefixing these values with their varint-encoded size, recursively. This was originally inspired by Alec Thomas's binary package, but I've reworked the serialization format and improved the performance and size. Here's a few notable features/goals of this `binary` package:
 * Zero-allocation encoding. I'm hoping to make the encoding to be as fast as possible, simply writing binary to the `io.Writer` without unncessary allocations.
 * Support for `maps`, `arrays`, `slices`, `structs`, primitive and nested types.
 * This is essentially a `json.Marshal` and `json.Unmarshal` drop-in replacement, I wanted this package to be simple to use and leverage the power of `reflect` package of golang.
 * The `ints` and `uints` are encoded using `varint`, making the payload small as possible.
 * Fast-paths encoding and decoding of `[]byte`, as I've designed this package to be used for inter-broker message encoding for [emitter](https://github.com/emitter-io/emitter).
 * Support for custom `BinaryMarshaler` and `BinaryUnmarshaler` for tighter packing control and built-in types such as `time.Time`.



# Usage
To serialize a message, simply `Marshal`:
```
v := &message{
    Name:      "Roman",
    Timestamp: 1242345235,
    Payload:   []byte("hi"),
    Ssid:      []uint32{1, 2, 3},
}

encoded, err := binary.Marshal(v)
```

To deserialize, `Unmarshal`:
```
var v message
err := binary.Unmarshal(encoded, &v)
```

# Disclaimer

This is not intended as a replacement for JSON or protobuf, this codec does not maintain any versioning or compatibility - and not intended to become one. The goal of this binary codec is to efficiently exchange binary data of known format between systems where you control both ends and both of them are written in Go.

