package perf

import (
	"github.com/kelindar/process"
)

type Stats struct {
	CPU           float64
	MemoryPrivate int64
	MemoryVirtual int64
}

func Sample() *Stats {
	stats := new(Stats)
	process.ProcUsage(&stats.CPU, &stats.MemoryPrivate, &stats.MemoryVirtual)

	return stats
}
