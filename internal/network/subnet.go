package network

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/alexflint/go-filemutex"
	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/internal/utils"
	"gitlab.com/amit-yuval/locker/pkg/io"
)

const (
	lockFile    = "/tmp/locker.lock"
	subnetsFile = "/var/lib/locker/subnets"
)

type subnet struct {
	addresses []int
	curAddr   int
}

func newSubnet(subnetAddr []int) *subnet {
	return &subnet{
		addresses: subnetAddr,
		curAddr:   1,
	}
}

func (s *subnet) toString() string {
	return utils.SplitToString(s.addresses, ".")
}

func (s *subnet) write() error {
	return io.WriteToFile(subnetsFile, os.O_APPEND|os.O_WRONLY, s.toString()+"\n")
}

func (s *subnet) destruct() error {
	m, err := filemutex.New(lockFile)
	if err != nil {
		return errors.Wrapf(err, "couldn't open lock %v", lockFile)
	}
	m.Lock()

	dat, err := ioutil.ReadFile(subnetsFile)
	if err != nil {
		m.Unlock()
		return errors.Wrap(err, "couldn't open subnets file")
	}
	contentStr := string(dat)
	contentStr = strings.Replace(contentStr, s.toString()+"\n", "", 1)
	io.WriteToFile(subnetsFile, os.O_TRUNC|os.O_WRONLY, contentStr)

	m.Unlock()

	return nil
}

func generateValidSubnet() (*subnet, error) {
	m, err := filemutex.New(lockFile)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't open lock %v", lockFile)
	}
	m.Lock()

	valid := false
	var subnetAddr []int
	for !valid {
		subnetAddr = generateSubnet()
		if validSubnet(subnetAddr) {
			valid = true
		}
	}
	subnet := newSubnet(subnetAddr)
	subnet.write()

	m.Unlock()
	return subnet, nil
}

func generateSubnet() []int {
	subnet := make([]int, 4)
	for i := 0; i < len(subnet)-1; i++ {
		subnet[i] = utils.RandRange(255, 0)
	}
	subnet[3] = 0
	return subnet
}

func subnetExists(subnet []int) bool {
	subnetStr := utils.SplitToString(subnet, ".")
	return io.FileContainsLine(subnetsFile, subnetStr)
}

func validSubnet(subnet []int) bool {
	return subnet[0] != 127 &&
		subnet[0] != 0 &&
		subnet[0] != 255 &&
		subnet[0] != 192 &&
		!subnetExists(subnet)
}

func createSubnet() []int {
	return nil
}
