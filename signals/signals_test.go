package signals_test

import (
	"syscall"
	"testing"

	"github.com/CodeLieutenant/utils/signals"
	"github.com/stretchr/testify/require"
)

func TestInvalidSignalError(t *testing.T) {
	// Arrange
	t.Parallel()
	assert := require.New(t)
	signal := "TEST"

	// Act
	sig, err := signals.Get(signal)

	// Assert
	assert.Nil(sig)
	assert.Error(err)
	assert.EqualError(err, "Cannot find signal TEST")
}

func TestGetSignalSuccess(t *testing.T) {
	// Arrange
	assert := require.New(t)
	signal := "SIGHUP"
	// Act
	sig, err := signals.Get(signal)

	// Assert
	assert.NotNil(sig)
	assert.NoError(err)
	assert.Equal(syscall.SIGHUP, sig)
}
