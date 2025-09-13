//nolint:forcetypeassert
package utils_test

import (
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/CodeLieutenant/utils"
	"github.com/stretchr/testify/require"
)

func TestGetEnvVariables(t *testing.T) {
	t.Parallel()
	assertRoot := require.New(t)

	type Test struct {
		expected any
		assert   func(utils.Env, Test)
		key      string
		value    string
	}

	tests := []Test{
		{key: "STRINGS", value: "string_value", expected: "string_value", assert: func(provider utils.Env, test Test) {
			val := utils.GetEnv(provider, test.key, "")
			assertRoot.Equal(test.expected.(string), val) //nolint:forcetypeassert
		}},
		{key: "INT", value: "10", expected: 10, assert: func(provider utils.Env, test Test) {
			val := utils.GetIntEnv[int](provider, test.key, -1)
			assertRoot.Equal(test.expected.(int), val) //nolint:forcetypeassert
		}},
		// Floats
		{key: "FLOAT32", value: "1.5", expected: float32(1.5), assert: func(p utils.Env, tc Test) {
			assertRoot.InEpsilon(tc.expected.(float32), utils.GetFloatEnv[float32](p, tc.key, 0), 0.001)
		}},
		{key: "FLOAT64", value: "3.14159", expected: 3.14159, assert: func(p utils.Env, tc Test) {
			assertRoot.InEpsilon(tc.expected.(float64), utils.GetFloatEnv[float64](p, tc.key, 0), 0.0001)
		}},
		// Unsigned
		{key: "UINT", value: "42", expected: uint(42), assert: func(p utils.Env, tc Test) { assertRoot.Equal(tc.expected.(uint), utils.GetUintEnv[uint](p, tc.key, 0)) }},
		{key: "UINT16", value: "65000", expected: uint16(65000), assert: func(p utils.Env, tc Test) {
			assertRoot.Equal(tc.expected.(uint16), utils.GetUintEnv[uint16](p, tc.key, 0))
		}},
		// Bool
		{key: "BOOL_TRUE", value: "true", expected: true, assert: func(p utils.Env, tc Test) { assertRoot.Equal(tc.expected.(bool), utils.GetBoolEnv(p, tc.key, false)) }},
		{key: "BOOL_FALSE", value: "false", expected: false, assert: func(p utils.Env, tc Test) { assertRoot.Equal(tc.expected.(bool), utils.GetBoolEnv(p, tc.key, true)) }},
		// Duration direct parse
		{key: "DURATION_PARSE", value: "250ms", expected: 250 * time.Millisecond, assert: func(p utils.Env, tc Test) {
			assertRoot.Equal(tc.expected.(time.Duration), utils.GetDurationEnv(p, tc.key, time.Second))
		}},
		// Duration seconds fallback path
		{key: "DURATION_SECONDS", value: "10", expected: 10 * time.Second, assert: func(p utils.Env, tc Test) {
			assertRoot.Equal(tc.expected.(time.Duration), utils.GetDurationEnv(p, tc.key, time.Second))
		}},
	}

	for _, test := range tests {
		test := test // capture
		t.Run("Test_"+test.key, func(t *testing.T) {
			provider := utils.NewTestEnv(t)
			t.Parallel()
			provider.Set(test.key, test.value)
			test.assert(provider, test)
		})
	}
}

func TestGetEnvVariables_Defaults(t *testing.T) {
	// Ensure defaults are returned when keys are missing.
	t.Parallel()
	prov := utils.NewTestEnv(t)

	require.Equal(t, "def", utils.GetEnv(prov, "MISSING_STRING", "def"))
	require.Equal(t, int64(99), utils.GetIntEnv[int64](prov, "MISSING_INT64", 99))
	require.Equal(t, uint32(77), utils.GetUintEnv[uint32](prov, "MISSING_UINT32", 77))
	require.InEpsilon(t, 1.25, utils.GetFloatEnv[float64](prov, "MISSING_FLOAT64", 1.25), 0.001)
	require.True(t, utils.GetBoolEnv(prov, "MISSING_BOOL", true))
	require.Equal(t, 5*time.Second, utils.GetDurationEnv(prov, "MISSING_DURATION", 5*time.Second))
}

func TestGetEnvVariables_Panics(t *testing.T) {
	t.Parallel()
	prov := utils.NewTestEnv(t)

	prov.Set("INT_BAD", "NaN")
	require.Panics(t, func() { _ = utils.GetIntEnv[int](prov, "INT_BAD", 0) })

	prov.Set("UINT_BAD", "-1")
	require.Panics(t, func() { _ = utils.GetUintEnv[uint](prov, "UINT_BAD", 0) })

	prov.Set("FLOAT_BAD", "abc")
	require.Panics(t, func() { _ = utils.GetFloatEnv[float64](prov, "FLOAT_BAD", 0) })

	prov.Set("BOOL_BAD", "maybe")
	require.Panics(t, func() { _ = utils.GetBoolEnv(prov, "BOOL_BAD", false) })

	prov.Set("DURATION_BAD", "nonsense")
	require.Panics(t, func() { _ = utils.GetDurationEnv(prov, "DURATION_BAD", time.Second) })
}

func TestGetInt_OverflowInt8(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("INT8_OVERFLOW", "128") // > int8 max 127
	require.Panics(t, func() { _ = utils.GetIntEnv[int8](p, "INT8_OVERFLOW", 0) })
}

func TestGetInt_OverflowInt16(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("INT16_OVERFLOW", "40000") // > int16 max 32767
	require.Panics(t, func() { _ = utils.GetIntEnv[int16](p, "INT16_OVERFLOW", 0) })
}

func TestGetUint_NegativeValue(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("UINT_NEG", "-5")
	require.Panics(t, func() { _ = utils.GetUintEnv[uint](p, "UINT_NEG", 0) })
}

func TestGetUint_OverflowUint8(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("U8_OVERFLOW", "256") // > 255
	require.Panics(t, func() { _ = utils.GetUintEnv[uint8](p, "U8_OVERFLOW", 0) })
}

func TestGetFloat_InvalidEmpty(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("F_EMPTY", "")
	require.Panics(t, func() { _ = utils.GetFloatEnv[float64](p, "F_EMPTY", 1) })
}

func TestGetBool_InvalidValues(t *testing.T) {
	t.Parallel()
	cases := []string{"ttrue", "1true", "yes", ""}
	for _, v := range cases {
		v := v
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			p := utils.NewTestEnv(t)
			p.Set("BOOL_BAD_GENERIC", v)
			require.Panics(t, func() { _ = utils.GetBoolEnv(p, "BOOL_BAD_GENERIC", false) })
		})
	}
}

func TestGetDurationEnv_FallbackLargeSeconds(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("DUR_LARGE_SECS", "7200") // 2h fallback path
	got := utils.GetDurationEnv(p, "DUR_LARGE_SECS", time.Second)
	require.Equal(t, 2*time.Hour, got)
}

func TestGetDurationEnv_EmptyPanics(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("DUR_EMPTY", "")
	require.Panics(t, func() { _ = utils.GetDurationEnv(p, "DUR_EMPTY", time.Minute) })
}

func TestGetEnv_EmptyStringVsMissing(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	// Missing returns default
	require.Equal(t, "fallback", utils.GetEnv(p, "MISSING_KEY_FOR_EMPTY", "fallback"))
	// Explicit empty should return empty (distinguish from default)
	p.Set("EXPLICIT_EMPTY", "")
	require.Empty(t, utils.GetEnv(p, "EXPLICIT_EMPTY", "fallback"))
}

func TestNumericExtremes(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("INT64_MAX", "9223372036854775807")
	p.Set("UINT64_MAX", "18446744073709551615")
	p.Set("FLOAT_MAX", "1.7976931348623157e+308") // ~math.MaxFloat64

	require.Equal(t, int64(math.MaxInt64), utils.GetIntEnv[int64](p, "INT64_MAX", 0))
	require.Equal(t, uint64(math.MaxUint64), utils.GetUintEnv[uint64](p, "UINT64_MAX", 0))
	require.InEpsilon(t, math.MaxFloat64, utils.GetFloatEnv[float64](p, "FLOAT_MAX", 0), 0.0001)
}

func TestParallelIsolation(t *testing.T) {
	t.Parallel()
	const runs = 25
	for i := 0; i < runs; i++ {
		i := i
		t.Run("iso_"+time.Duration(i).String(), func(t *testing.T) {
			t.Parallel()
			p := utils.NewTestEnv(t)
			key := "K" + strconv.Itoa(i)
			val := strconv.Itoa(100 + i)
			p.Set(key, val)
			require.Equal(t, 100+i, utils.GetIntEnv[int](p, key, 0))
		})
	}
}

func TestGetAllIntTypesParsing(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("INT_STD", "123")
	p.Set("INT8_VAL", "-12")
	p.Set("INT16_VAL", "32000")
	p.Set("INT32_VAL", "2000000000")
	p.Set("INT64_VAL", "9223372036854775807")

	require.Equal(t, 123, utils.GetIntEnv[int](p, "INT_STD", 0))
	require.Equal(t, int8(-12), utils.GetIntEnv[int8](p, "INT8_VAL", 0))
	require.Equal(t, int16(32000), utils.GetIntEnv[int16](p, "INT16_VAL", 0))
	require.Equal(t, int32(2000000000), utils.GetIntEnv[int32](p, "INT32_VAL", 0))
	require.Equal(t, int64(math.MaxInt64), utils.GetIntEnv[int64](p, "INT64_VAL", 0))
}

func TestGetAllIntTypesOverflow(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("INT8_OVER", "128")                  // > 127
	p.Set("INT16_OVER", "40000")               // > 32767
	p.Set("INT32_OVER", "2147483648")          // > int32 max
	p.Set("INT64_OVER", "9223372036854775808") // > int64 max
	p.Set("INT_STD_OVER", "9223372036854775808")

	require.Panics(t, func() { _ = utils.GetIntEnv[int8](p, "INT8_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetIntEnv[int16](p, "INT16_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetIntEnv[int32](p, "INT32_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetIntEnv[int64](p, "INT64_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetIntEnv[int](p, "INT_STD_OVER", 0) })
}

func TestGetAllUintTypesParsing(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("UINT_STD", "123")
	p.Set("UINT8_VAL", "255")
	p.Set("UINT16_VAL", "65000")
	p.Set("UINT32_VAL", "4294967295")
	p.Set("UINT64_VAL", "18446744073709551615")

	require.Equal(t, uint(123), utils.GetUintEnv[uint](p, "UINT_STD", 0))
	require.Equal(t, uint8(255), utils.GetUintEnv[uint8](p, "UINT8_VAL", 0))
	require.Equal(t, uint16(65000), utils.GetUintEnv[uint16](p, "UINT16_VAL", 0))
	require.Equal(t, uint32(4294967295), utils.GetUintEnv[uint32](p, "UINT32_VAL", 0))
	require.Equal(t, uint64(math.MaxUint64), utils.GetUintEnv[uint64](p, "UINT64_VAL", 0))
}

func TestGetAllUintTypesOverflow(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("UINT8_OVER", "256")
	p.Set("UINT16_OVER", "70000")
	p.Set("UINT32_OVER", "4294967296")
	p.Set("UINT64_OVER", "18446744073709551616")
	p.Set("UINT_STD_OVER", "18446744073709551616")

	require.Panics(t, func() {
		_ = utils.GetUintEnv[uint8](p, "UINT8_OVER", 0)
	})
	require.Panics(t, func() { _ = utils.GetUintEnv[uint16](p, "UINT16_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetUintEnv[uint32](p, "UINT32_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetUintEnv[uint64](p, "UINT64_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetUintEnv[uint](p, "UINT_STD_OVER", 0) })
}

func TestGetFloatTypesParsing(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("FLOAT32_POS", "3.25")
	p.Set("FLOAT32_NEG", "-1.5")
	p.Set("FLOAT64_POS", "6.02214076e23")
	p.Set("FLOAT64_NEG", "-2.718281828")

	require.InEpsilon(t, float32(3.25), utils.GetFloatEnv[float32](p, "FLOAT32_POS", 0), 0.0001)
	require.InEpsilon(t, float32(-1.5), utils.GetFloatEnv[float32](p, "FLOAT32_NEG", 0), 0.0001)
	require.InEpsilon(t, 6.02214076e23, utils.GetFloatEnv[float64](p, "FLOAT64_POS", 0), 0.0001)
	require.InEpsilon(t, -2.718281828, utils.GetFloatEnv[float64](p, "FLOAT64_NEG", 0), 0.00001)
}

func TestGetFloatTypesOverflow(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	// Values that overflow respective sizes -> strconv.ParseFloat returns error -> panic
	p.Set("F32_OVER", "3.5e39") // > ~3.4e38
	p.Set("F64_OVER", "1e5000") // huge

	require.Panics(t, func() { _ = utils.GetFloatEnv[float32](p, "F32_OVER", 0) })
	require.Panics(t, func() { _ = utils.GetFloatEnv[float64](p, "F64_OVER", 0) })
}

func TestGetBoolVariants(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)
	p.Set("BOOL_TRUE_UPPER", "TRUE")
	p.Set("BOOL_FALSE_MIXED", "False")
	p.Set("BOOL_ONE", "1")
	p.Set("BOOL_ZERO", "0")
	p.Set("BOOL_T", "t")
	p.Set("BOOL_F", "f")

	require.True(t, utils.GetBoolEnv(p, "BOOL_TRUE_UPPER", false))
	require.False(t, utils.GetBoolEnv(p, "BOOL_FALSE_MIXED", true))
	require.True(t, utils.GetBoolEnv(p, "BOOL_ONE", false))
	require.False(t, utils.GetBoolEnv(p, "BOOL_ZERO", true))
	require.True(t, utils.GetBoolEnv(p, "BOOL_T", false))
	require.False(t, utils.GetBoolEnv(p, "BOOL_F", true))

	// Missing key -> default
	require.True(t, utils.GetBoolEnv(p, "BOOL_MISSING", true))
}

func TestGetStringsEnv(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)

	// Test with multiple comma-separated values
	p.Set("STRINGS_MULTI", "value1,value2,value3")
	result := utils.GetStringsEnv(p, "STRINGS_MULTI", []string{"default"})
	require.Equal(t, []string{"value1", "value2", "value3"}, result)

	// Test with spaces around values
	p.Set("STRINGS_SPACES", " value1 , value2 , value3 ")
	result = utils.GetStringsEnv(p, "STRINGS_SPACES", []string{"default"})
	require.Equal(t, []string{"value1", "value2", "value3"}, result)

	// Test with empty values mixed in
	p.Set("STRINGS_EMPTY_MIXED", "value1,,value2, ,value3")
	result = utils.GetStringsEnv(p, "STRINGS_EMPTY_MIXED", []string{"default"})
	require.Equal(t, []string{"value1", "value2", "value3"}, result)

	// Test with single value
	p.Set("STRINGS_SINGLE", "singlevalue")
	result = utils.GetStringsEnv(p, "STRINGS_SINGLE", []string{"default"})
	require.Equal(t, []string{"singlevalue"}, result)

	// Test with empty string
	p.Set("STRINGS_EMPTY", "")
	result = utils.GetStringsEnv(p, "STRINGS_EMPTY", []string{"default"})
	require.Equal(t, []string{}, result)

	// Test with only spaces and commas
	p.Set("STRINGS_ONLY_SPACES", " , , ")
	result = utils.GetStringsEnv(p, "STRINGS_ONLY_SPACES", []string{"default"})
	require.Equal(t, []string{}, result)

	// Test missing key returns default
	result = utils.GetStringsEnv(p, "STRINGS_MISSING", []string{"default1", "default2"})
	require.Equal(t, []string{"default1", "default2"}, result)
}

func TestGetKeyEnv(t *testing.T) {
	t.Parallel()
	p := utils.NewTestEnv(t)

	// Test with valid hex key (32 bytes when decoded)
	validHexKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes hex
	p.Set("CRYPTO_KEY_HEX", validHexKey)
	defaultKey := []byte("defaultkey")
	result := utils.GetKeyEnv(p, "CRYPTO_KEY_HEX", defaultKey)
	require.NotEqual(t, defaultKey, result)
	require.Len(t, result, 32)

	// Test with valid base64 key with prefix
	validBase64Key := "base64:dGVzdGtleTE2Ynl0ZXNmb3J0ZXN0aW5nMTIzNA==" // "testkey16bytesffortesting1234" base64 encoded with prefix
	p.Set("CRYPTO_KEY_B64", validBase64Key)
	result = utils.GetKeyEnv(p, "CRYPTO_KEY_B64", defaultKey)
	require.NotEqual(t, defaultKey, result)
	require.NotEmpty(t, result)

	// Test with empty key returns default
	p.Set("EMPTY_KEY", "")
	result = utils.GetKeyEnv(p, "EMPTY_KEY", defaultKey)
	require.Equal(t, defaultKey, result)

	// Test missing key returns default
	result = utils.GetKeyEnv(p, "MISSING_KEY", defaultKey)
	require.Equal(t, defaultKey, result)

	// Test with invalid hex format panics
	p.Set("INVALID_HEX_KEY", "invalid-hex-format")
	require.Panics(t, func() {
		_ = utils.GetKeyEnv(p, "INVALID_HEX_KEY", defaultKey)
	})

	// Test with invalid base64 format panics
	p.Set("INVALID_B64_KEY", "base64:invalid-base64!")
	require.Panics(t, func() {
		_ = utils.GetKeyEnv(p, "INVALID_B64_KEY", defaultKey)
	})
}

func TestNewEnv(t *testing.T) {
	t.Parallel()

	// Test NewEnv with dotenv disabled
	env := utils.NewEnv(true)
	require.NotNil(t, env.EnvProvider)

	// Test NewEnv with dotenv enabled (will try to load .env file)
	// This might panic if .env file doesn't exist, which is expected behavior
	defer func() {
		if r := recover(); r != nil {
			// Expected panic if .env file doesn't exist
			require.Contains(t, r.(string), "dotenv file not found")
		}
	}()
	env = utils.NewEnv(false)
	require.NotNil(t, env.EnvProvider)
}

func TestLoadDotEnv_Success(t *testing.T) {
	t.Parallel()

	// Create a temporary .env file for testing successful loading
	envContent := "TEST_VAR=test_value\n"
	err := os.WriteFile(".env", []byte(envContent), 0o644)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(".env")
	}()

	// This should not panic and should load successfully
	require.NotPanics(t, func() {
		utils.LoadDotEnv()
	})

	// Verify the variable was loaded
	value := os.Getenv("TEST_VAR")
	require.Equal(t, "test_value", value)

	// Clean up
	err = os.Unsetenv("TEST_VAR")
	if err != nil {
		t.Fatalf("failed to unset env var: %v", err)
	}
}

func TestLoadDotEnv_LoadError(t *testing.T) {
	t.Parallel()

	// Create an invalid .env file that will cause godotenv.Load to fail
	invalidEnvContent := "INVALID_LINE_WITHOUT_EQUALS\n"
	err := os.WriteFile(".env", []byte(invalidEnvContent), 0o644)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(".env")
	}()

	// This should not panic but should log an error and return
	require.NotPanics(t, func() {
		utils.LoadDotEnv()
	})
}

//nolint:paralleltest
func TestOsEnvProvider(t *testing.T) {
	provider := utils.OSEnvProvider{}

	// Test Set and Get with actual OS environment
	testKey := "TEST_ENV_KEY"
	testValue := "test_value_12345"

	// Clean up any existing value
	originalValue := os.Getenv(testKey)
	defer func() {
		if originalValue != "" {
			t.Setenv(testKey, originalValue)
		} else {
			t.Setenv(testKey, "")
		}
	}()

	// Test Set
	provider.Set(testKey, testValue)

	// Test Get - should return the value we just set
	result, exists := provider.Get(testKey)
	require.True(t, exists)
	require.Equal(t, testValue, result)

	// Test Get with non-existent key
	result, exists = provider.Get("NON_EXISTENT_KEY_12345")
	require.False(t, exists)
	require.Empty(t, result)

	// Test Set with empty value
	provider.Set(testKey, "")
	result, exists = provider.Get(testKey)
	require.True(t, exists)
	require.Empty(t, result)
}

func TestOsEnvProvider_SetError(t *testing.T) {
	t.Parallel()

	provider := utils.OSEnvProvider{}

	// Test Set with invalid environment variable name that will cause os.Setenv to fail
	// Environment variable names with null bytes are invalid and will cause os.Setenv to fail
	invalidKey := "INVALID\x00KEY"

	require.Panics(t, func() {
		provider.Set(invalidKey, "test_value")
	})
}
