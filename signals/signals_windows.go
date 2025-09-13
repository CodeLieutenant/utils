//go:build windows

package signals

import (
	"os"
	"syscall"
)

var signals = map[string]os.Signal{
	"SIGHUP":  syscall.SIGHUP,
	"SIGINT":  syscall.SIGINT,
	"SIGQUIT": syscall.SIGQUIT,
	"SIGKILL": syscall.SIGKILL,
	"SIGALRM": syscall.SIGALRM,
	"SIGTERM": syscall.SIGTERM,
}
