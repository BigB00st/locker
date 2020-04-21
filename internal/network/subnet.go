package network

import (
	"github.com/alexflint/go-filemutex"
	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/internal/utils"
)

const (
	lockFile = "/tmp/locker.lock"
)

type subnet struct {
	addresses []int
	curAddr   int
}

func newSubnet() *subnet {
	return &subnet{
		addresses: make([]int, 4),
		curAddr:   1,
	}
}

func (s subnet) toString() string {
	return utils.SplitToString(s.addresses, ".")
}

func getSubnet() error {
	m, err := filemutex.New(lockFile)
	if err != nil {
		return errors.Wrapf(err, "couldn't create lock %v", lockFile)
	}

	m.Lock() // Will block until lock can be acquired

	// Code here is protected by the mutex

	m.Unlock()

	return nil
}

func generateSubnet() []int {
	return nil
}

func subnetExists() bool {
	return true
}

func isSubnetValid() bool {
	return true
}

func createSubnet() []int {
	return nil
}
