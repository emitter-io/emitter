package hyperloglog

import (
	"math"
)

type reg uint8
type tailcuts []reg

type registers struct {
	tailcuts
	nz uint32
}

func (r *reg) set(offset, val uint8) bool {
	var isZero bool
	if offset == 0 {
		isZero = uint8((*r)>>4) == 0
		tmpVal := uint8((*r) << 4 >> 4)
		*r = reg(tmpVal | (val << 4))
	} else {
		isZero = uint8((*r)<<4>>4) == 0
		tmpVal := uint8((*r) >> 4)
		*r = reg(tmpVal<<4 | val)
	}
	return isZero
}

func (r *reg) get(offset uint8) uint8 {
	if offset == 0 {
		return uint8((*r) >> 4)
	}
	return uint8((*r) << 4 >> 4)
}

func newRegisters(size uint32) *registers {
	return &registers{
		tailcuts: make(tailcuts, size/2),
		nz:       size,
	}
}

func (rs *registers) clone() *registers {
	if rs == nil {
		return nil
	}
	tc := make([]reg, len(rs.tailcuts))
	copy(tc, rs.tailcuts)
	return &registers{
		tailcuts: tc,
		nz:       rs.nz,
	}
}

func (rs *registers) rebase(delta uint8) {
	nz := uint32(len(rs.tailcuts)) * 2
	for i := range rs.tailcuts {
		val := rs.tailcuts[i].get(0)
		if val >= delta {
			rs.tailcuts[i].set(0, val-delta)
			if val-delta > 0 {
				nz--
			}
		}
		val = rs.tailcuts[i].get(1)
		if val >= delta {
			rs.tailcuts[i].set(1, val-delta)
			if val-delta > 0 {
				nz--
			}
		}
	}
	rs.nz = nz
}

func (rs *registers) set(i uint32, val uint8) {
	offset, index := uint8(i%2), i/2
	if rs.tailcuts[index].set(offset, val) {
		rs.nz--
	}
}

func (rs *registers) get(i uint32) uint8 {
	offset, index := uint8(i%2), i/2
	return rs.tailcuts[index].get(offset)
}

func (rs *registers) sumAndZeros(base uint8) (res, ez float64) {
	for _, r := range rs.tailcuts {
		v1 := float64(base + r.get(0))
		if v1 == 0 {
			ez++
		}
		res += 1.0 / math.Pow(2.0, v1)
		v2 := float64(base + r.get(0))
		if v2 == 0 {
			ez++
		}
		res += 1.0 / math.Pow(2.0, float64(base+r.get(1)))

	}
	rs.nz = uint32(ez)
	return res, ez
}

func (rs *registers) min() uint8 {
	if rs.nz > 0 {
		return 0
	}
	min := uint8(math.MaxUint8)
	for _, r := range rs.tailcuts {
		if val := uint8(r << 4 >> 4); val < min {
			min = val
		}
		if val := uint8(r >> 4); val < min {
			min = val
		}
		if min == 0 {
			break
		}
	}
	return min
}
