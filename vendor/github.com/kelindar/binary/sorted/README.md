# Sorted integer slices for small payloads

This sub-package contains a set of sorted sclices for **minimising the payload**. This can be useful in certain situations where you can sort a slice and send it to through the wire in the sorted format. This is essentially a trade-off between CPU and network bandwith.

# Usage
This is a drop-in type, so simply use one of the types available in the package (`Bools`, `Int32s`, `Uint64s` ...) and `Marshal` or `Unmarshal` using the binary package.
```
// Marshal some numbers
v := sorted.Int32s{4, 5, 6, 1, 2, 3}
encoded, err := binary.Marshal(&v)

// Unmarshal the numbers
var o sorted.Int32s
err = binary.Unmarshal(encoded, &o)
```
