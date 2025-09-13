package utils

import (
	"math"
	"slices"
	"strconv"
)

type MemorySize uint64

const (
	KiB MemorySize = 1 << 10
	MiB MemorySize = 1 << 20
	GiB MemorySize = 1 << 30
	TiB MemorySize = 1 << 40
)

func (m MemorySize) String() string {
	switch {
	case m >= TiB:
		return FormatMemorySize(m, TiB, "TiB")
	case m >= GiB:
		return FormatMemorySize(m, GiB, "GiB")
	case m >= MiB:
		return FormatMemorySize(m, MiB, "MiB")
	case m >= KiB:
		return FormatMemorySize(m, KiB, "KiB")
	default:
		return FormatMemorySize(m, 1, "B")
	}
}

var largestFloatSize = len(strconv.FormatFloat(math.MaxFloat64, 'f', 0, 64))

func FormatMemorySize(m MemorySize, unit MemorySize, suffix string) string {
	bytes := make([]byte, 0, len(suffix)+largestFloatSize)

	bytes = strconv.AppendFloat(bytes, float64(m)/float64(unit), 'f', 0, 64)
	bytes = append(bytes, suffix...)

	return UnsafeString(slices.Clip(bytes))
}
