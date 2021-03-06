package network

import (
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/amit-yuval/locker/internal/utils"
	"gitlab.com/amit-yuval/locker/pkg/io"

	"github.com/pkg/errors"
	"github.com/vishvananda/netns"
)

const (
	subnetMaskBytes     = 4
	subnetLogicOne      = 255
	bitsInByte          = 8
	ipRouteDefaultIndex = 0
	ipRouteNameIndex    = 4
	netnsDirectory      = "/var/run/netns/"
	interfaceNameLen    = 11
	interfacePrefix     = "veth"
	nsNameLen           = 10
	nsPrefix            = "ns-"
)

// NetConfig holds the network namespace name of the container
type NetConfig struct {
	sub                                          *subnet
	nsName, masqueradeIp, netInterface, vethName string
	prevNs                                       netns.NsHandle
}

// CreateConnectivity creates isolated network connectivity for the container
func CreateConnectivity() (NetConfig, error) {
	netConfig := NetConfig{}

	netInterface, err := connectedInterfaceName()
	if err != nil {
		return netConfig, err
	}

	netConfig.sub, err = generateValidSubnet()
	if err != nil {
		return netConfig, err
	}

	nsName, err := utils.GetUnique(nsPrefix, nsNameLen, utils.CreateUuid, netNsExists)
	if err != nil {
		return netConfig, err
	}
	netConfig.nsName = nsName

	vethName, err := utils.GetUnique(interfacePrefix, nsNameLen, utils.CreateUuid, netInterfaceExists)
	if err != nil {
		return netConfig, err
	}
	vethPeerName := vethName + "-p"
	vethIp := netConfig.sub.nextIp()
	vethCIDR := vethIp + "/24"
	vethPeerCIDR := netConfig.sub.nextIp() + "/24"
	loopback := "lo"
	masqueradeIp := netConfig.sub.toString() + "/255.255.255.0"
	curNs, err := netns.Get()
	if err != nil {
		return netConfig, errors.Wrap(err, "couldn't get current ns")
	}

	netConfig.masqueradeIp = masqueradeIp
	netConfig.netInterface = netInterface
	netConfig.vethName = vethName
	netConfig.prevNs = curNs

	// create network namespace
	if err := addNetNs(nsName); err != nil {
		return netConfig, err
	}

	// create veth pair
	if err := addVethPair(vethName, vethPeerName); err != nil {
		return netConfig, err
	}

	// assign peer to namespace
	if err := assignVethToNs(vethPeerName, nsName); err != nil {
		return netConfig, err
	}

	// setup ipv4 of veth
	if err := addIp(vethCIDR, vethName); err != nil {
		return netConfig, err
	}
	if err := setInterfaceUp(vethName); err != nil {
		return netConfig, err
	}

	// setup ipv4 of veth peer
	if err := addIpInsideNs(vethPeerCIDR, vethPeerName, nsName); err != nil {
		return netConfig, err
	}
	if err := setInterfaceUpInsideNs(vethPeerName, nsName); err != nil {
		return netConfig, err
	}
	if err := setInterfaceUpInsideNs(loopback, nsName); err != nil {
		return netConfig, err
	}

	// add default gateway inside namespace
	if err := addDefaultGateway(nsName, vethIp); err != nil {
		return netConfig, err
	}

	// enable ipv4 forwarding
	if err := enableIpv4Forwarding(); err != nil {
		return netConfig, err
	}

	// set rules to allow connectivity
	if err := setIptablesRules(masqueradeIp, netInterface, vethName); err != nil {
		return netConfig, err
	}

	if err := joinNsByName(nsName); err != nil {
		return netConfig, err
	}

	return netConfig, nil
}

// Cleanup deletes the created network namespace, and updates subnets file
func (c *NetConfig) Cleanup() {
	netns.Set(c.prevNs)
	if c.nsName != "" {
		deleteNetNs(c.nsName)
		clearIptablesRules(c.masqueradeIp, c.netInterface, c.vethName)
	}
	if c.sub != nil {
		c.sub.destruct()
	}
}

// joinNsByName gets file descriptor of requested network namespace, calls setNs with fd
func joinNsByName(nsName string) error {
	nsHandle, err := netns.GetFromName(nsName)
	if err != nil {
		return errors.Wrapf(err, "couldn't get fd of network namespace %q", nsName)
	}
	return netns.Set(nsHandle)
}

// addNetNs creates a network namespace of name nsName
func addNetNs(nsName string) error {
	if err := exec.Command("ip", "netns", "add", nsName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add new network namespace %q", nsName)
	}
	return nil
}

// Function deletes network namespace by the name of nsName
func deleteNetNs(nsName string) error {
	if err := exec.Command("ip", "netns", "delete", nsName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add new network namespace %q", nsName)
	}
	return nil
}

// function return true if namespace exists
func netNsExists(nsName string) bool {
	return io.FileExists(filepath.Join(netnsDirectory, nsName))
}

// addVethPair adds a virtual ethernet pair
func addVethPair(vethName, vethPeerName string) error {

	if netInterfaceExists(vethName) {
		return nil
	}

	if err := exec.Command("ip", "link", "add", vethName, "type", "veth", "peer", "name", vethPeerName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add veth pair of %q and %q", vethName, vethPeerName)
	}
	return nil
}

// function return true if Veth pair exists
func netInterfaceExists(vethName string) bool {
	out, _ := io.CmdOut("ip", "link", "list")
	return strings.Contains(out, vethName+"@")
}

// assignVethToNs assigns vethName network namespace to nsName
func assignVethToNs(vethName, nsName string) error {
	if err := exec.Command("ip", "link", "set", vethName, "netns", nsName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't assign veth %v to ns %v", vethName, nsName)
	}
	return nil
}

// addIp adds given ip to veth
func addIp(Ip, vethName string) error {
	if err := exec.Command("ip", "addr", "add", Ip, "dev", vethName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add ip to veth %v", vethName)
	}
	return nil
}

// addIpInsideNs adds given ip to veth (inside network namespace nsName)
func addIpInsideNs(Ip, vethName, nsName string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", Ip, "dev", vethName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add ip to %v inside ns %v", vethName, nsName)
	}
	return nil
}

// setInterfaceUpInsideNs sets vethName up (inside network namespace nsName)
func setInterfaceUpInsideNs(vethName, nsName string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethName, "up").Run(); err != nil {
		return errors.Wrapf(err, "couldn't set interface %q up inside namespace %q", vethName, nsName)
	}
	return nil
}

// setInterfaceUp sets vethName up
func setInterfaceUp(interfaceName string) error {
	if err := exec.Command("ip", "link", "set", interfaceName, "up").Run(); err != nil {
		return errors.Wrapf(err, "couldn't set interface %q up", interfaceName)
	}
	return nil
}

// addDefaultGateway adds default gateway to network namespace
func addDefaultGateway(nsName, Ip string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "route", "add", "default", "via", Ip).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add default gateway to %v", nsName)
	}
	return nil
}

// enableIpv4Forwarding enables kernel ipv4 forwarding
func enableIpv4Forwarding() error {
	if err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run(); err != nil {
		return errors.Wrap(err, "couldn't enable ipv4 forwarding")
	}
	return nil
}

// connectedInterfaceName returns name of currently connected interface name
func connectedInterfaceName() (string, error) {
	out, _ := io.CmdOut("ip", "-4", "route", "ls")

	for _, line := range strings.Split(out, "\n") {
		words := strings.Split(line, " ")
		if words[ipRouteDefaultIndex] == "default" {
			return words[ipRouteNameIndex], nil
		}
	}
	return "", errors.New("Not connected to the internet")
}
