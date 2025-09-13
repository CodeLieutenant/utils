package utils_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/CodeLieutenant/utils"
	"github.com/stretchr/testify/require"
)

func TestParseKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    []byte
		expectError bool
	}{
		{
			name:        "Empty key",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Base64 prefixed key",
			input:       "base64:HaFPpa+lXse9vfHgvVR56ij79la6qOXk7u6bg/RYqfo=",
			expected:    mustDecode(base64.StdEncoding.DecodeString("HaFPpa+lXse9vfHgvVR56ij79la6qOXk7u6bg/RYqfo=")),
			expectError: false,
		},
		{
			name:        "Hex key",
			input:       "1da14fa5afa55ec7bdbdf1e0bd5479ea28fbf656baa8e5e4eeee9b83f458a9fa",
			expected:    mustDecode(hex.DecodeString("1da14fa5afa55ec7bdbdf1e0bd5479ea28fbf656baa8e5e4eeee9b83f458a9fa")),
			expectError: false,
		},
		{
			name:        "Invalid base64 key",
			input:       "base64:invalid!base64",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid hex key",
			input:       "invalid!hex",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := utils.ParseKey(tt.input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// Helper function to handle decoding errors in test data setup
func mustDecode(data []byte, err error) []byte {
	if err != nil {
		panic(err)
	}

	return data
}
