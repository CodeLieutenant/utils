package signals

import (
	"errors"
	"os"
)

func Get(s string) (os.Signal, error) {
	val, ok := signals[s]

	if !ok {
		return nil, errors.New("Cannot find signal " + s)
	}

	return val, nil
}

func MustGet(s string) os.Signal {
	val, err := Get(s)
	if err != nil {
		panic(err)
	}

	return val
}
