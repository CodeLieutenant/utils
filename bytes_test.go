package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CodeLieutenant/utils"
)

func TestMemorySize_String(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	testCases := []struct {
		name     string
		size     utils.MemorySize
		expected string
	}{
		{
			name:     "bytes",
			size:     utils.MemorySize(512),
			expected: "512B",
		},
		{
			name:     "exactly 1 KiB",
			size:     utils.KiB,
			expected: "1KiB",
		},
		{
			name:     "multiple KiB",
			size:     utils.KiB * 1536, // 1.5 MiB
			expected: "2MiB",
		},
		{
			name:     "exactly 1 MiB",
			size:     utils.MiB,
			expected: "1MiB",
		},
		{
			name:     "multiple MiB",
			size:     utils.MiB * 512,
			expected: "512MiB",
		},
		{
			name:     "exactly 1 GiB",
			size:     utils.GiB,
			expected: "1GiB",
		},
		{
			name:     "multiple GiB",
			size:     utils.GiB * 4,
			expected: "4GiB",
		},
		{
			name:     "exactly 1 TiB",
			size:     utils.TiB,
			expected: "1TiB",
		},
		{
			name:     "multiple TiB",
			size:     utils.TiB * 2,
			expected: "2TiB",
		},
		{
			name:     "fractional KiB rounds down",
			size:     utils.KiB + 512, // 1.5 KiB
			expected: "2KiB",
		},
		{
			name:     "zero bytes",
			size:     utils.MemorySize(0),
			expected: "0B",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.size.String()
			req.Equal(tc.expected, result)
		})
	}
}

func TestFormatMemorySize(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	testCases := []struct {
		name     string
		size     utils.MemorySize
		unit     utils.MemorySize
		suffix   string
		expected string
	}{
		{
			name:     "bytes",
			size:     utils.MemorySize(1024),
			unit:     1,
			suffix:   "B",
			expected: "1024B",
		},
		{
			name:     "KiB",
			size:     utils.MemorySize(2048),
			unit:     utils.KiB,
			suffix:   "KiB",
			expected: "2KiB",
		},
		{
			name:     "MiB",
			size:     utils.MiB * 512,
			unit:     utils.MiB,
			suffix:   "MiB",
			expected: "512MiB",
		},
		{
			name:     "GiB",
			size:     utils.GiB * 8,
			unit:     utils.GiB,
			suffix:   "GiB",
			expected: "8GiB",
		},
		{
			name:     "TiB",
			size:     utils.TiB * 3,
			unit:     utils.TiB,
			suffix:   "TiB",
			expected: "3TiB",
		},
		{
			name:     "custom unit and suffix",
			size:     utils.MemorySize(1000),
			unit:     utils.MemorySize(100),
			suffix:   "CUSTOM",
			expected: "10CUSTOM",
		},
		{
			name:     "zero size",
			size:     utils.MemorySize(0),
			unit:     utils.MiB,
			suffix:   "MiB",
			expected: "0MiB",
		},
		{
			name:     "fractional result rounded to integer",
			size:     utils.MemorySize(1536), // 1.5 KiB
			unit:     utils.KiB,
			suffix:   "KiB",
			expected: "2KiB", // rounded to nearest integer
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.FormatMemorySize(tc.size, tc.unit, tc.suffix)
			req.Equal(tc.expected, result)
		})
	}
}

func TestMemorySize_Constants(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Test that the constants have expected values
	req.Equal(utils.MemorySize(1024), utils.KiB)
	req.Equal(utils.MemorySize(1024*1024), utils.MiB)
	req.Equal(utils.MemorySize(1024*1024*1024), utils.GiB)
	req.Equal(utils.MemorySize(1024*1024*1024*1024), utils.TiB)

	// Test relationships between constants
	req.Equal(utils.KiB*1024, utils.MiB)
	req.Equal(utils.MiB*1024, utils.GiB)
	req.Equal(utils.GiB*1024, utils.TiB)
}
