package network

import (
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/utils"
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

// SYS_SETNS syscall allows changing the namespace of the current process.
var SYS_SETNS = map[string]uintptr{
	"386":      346,
	"amd64":    308,
	"arm64":    268,
	"arm":      375,
	"mips":     4344,
	"mipsle":   4344,
	"mips64le": 4344,
	"ppc64":    350,
	"ppc64le":  350,
	"riscv64":  268,
	"s390x":    339,
}[runtime.GOARCH]

type NetConfig struct {
	nsName string
}

func CreateConnectivity() (NetConfig, error) {
	netConfig := NetConfig{}

	netInterface, err := connectedInterfaceName()
	if err != nil {
		return netConfig, err
	}

	nsName, err := utils.GetUnique(nsPrefix, nsNameLen, utils.CreateUuid, netNsExists)
	if err != nil {
		return netConfig, err
	}
	netConfig.nsName = nsName

	vethName, err := utils.GetUnique(interfacePrefix, nsNameLen, utils.CreateUuid, netNsExists)
	if err != nil {
		return netConfig, err
	}
	vethPeerName := vethName + "-p"
	vethIp := "10.200.1.1"
	vethCIDR := vethIp + "/24"
	vethPeerCIDR := "10.200.1.2/24"
	loopback := "lo"
	masqueradeIp := "10.200.1.0/255.255.255.0"

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

func (c *NetConfig) Cleanup() {
	if c.nsName != "" {
		deleteNetNs(c.nsName)
	}
}

func joinNsByName(nsName string) error {
	nsHandle, err := utils.GetFdFromPath(netnsDirectory + nsName)
	if err != nil {
		return errors.Wrapf(err, "couldn't get fd of network namespace %q", nsName)
	}
	return setNs(nsHandle, syscall.CLONE_NEWNET)
}

func setNs(nsHandle int, nsType int) error {
	if _, _, err := syscall.Syscall(SYS_SETNS, uintptr(nsHandle), uintptr(nsType), 0); err != 0 {
		return errors.Wrap(err, "couldn't set network namespace")
	}
	return nil
}

func addNetNs(nsName string) error {

	if netNsExists(nsName) {
		return nil
	}

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
	return utils.FileExists(filepath.Join(netnsDirectory, nsName))
}

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
	out, _ := utils.CmdOut("ip", "link", "list")
	return strings.Contains(out, vethName+"@")
}

func assignVethToNs(vethName, nsName string) error {
	if err := exec.Command("ip", "link", "set", vethName, "netns", nsName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't assign veth %v to ns %v", vethName, nsName)
	}
	return nil
}

func addIp(Ip, vethName string) error {
	if err := exec.Command("ip", "addr", "add", Ip, "dev", vethName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add ip to veth %v", vethName)
	}
	return nil
}

func addIpInsideNs(Ip, vethName, nsName string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", Ip, "dev", vethName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add ip to %v inside ns %v", vethName, nsName)
	}
	return nil
}

func setInterfaceUpInsideNs(vethName, nsName string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethName, "up").Run(); err != nil {
		return errors.Wrapf(err, "couldn't set interface %q up inside namespace %q", vethName, nsName)
	}
	return nil
}

func bridgeExists(bridgeName string) bool {
	out, _ := utils.CmdOut("ip", "link", "list", "type", "bridge")
	return strings.Contains(out, bridgeName)
}

func createBridge(bridgeName string) error {
	if bridgeExists(bridgeName) {
		return nil
	}

	if err := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge").Run(); err != nil {
		return errors.Wrapf(err, "couldn't create bridge %q", bridgeName)
	}
	return nil
}

func setInterfaceUp(interfaceName string) error {
	if err := exec.Command("ip", "link", "set", interfaceName, "up").Run(); err != nil {
		return errors.Wrapf(err, "couldn't set interface %q up", interfaceName)
	}
	return nil
}

func addInterfaceToBridge(interfaceName, bridgeName string) error {
	if err := exec.Command("ip", "link", "set", interfaceName, "master", bridgeName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add bridge %q to interface %q", bridgeName, interfaceName)
	}
	return nil
}

func addBridgeIp(bridgeName, Ip string) error {
	if err := exec.Command("ip", "addr", "add", Ip, "brd", "+", "dev", bridgeName).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add ip %q to bridge %q", Ip, bridgeName)
	}
	return nil
}

func addDefaultGateway(nsName, Ip string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "route", "add", "default", "via", Ip).Run(); err != nil {
		return errors.Wrapf(err, "couldn't add default gateway to %v", nsName)
	}
	return nil
}

func setIptablesRules(masqueradeIp, netInterface, vethName string) error {
	// Policy DROP by default.
	if err := exec.Command("iptables", "-P", "FORWARD", "DROP").Run(); err != nil {
		return errors.Wrap(err, "couldn't policy DROP by default")
	}

	// Flush forward rules,
	if err := exec.Command("iptables", "-F", "FORWARD").Run(); err != nil {
		return errors.Wrap(err, "couldn't flush forward rules")
	}

	// Flush nat rules.
	if err := exec.Command("iptables", "-t", "nat", "-F").Run(); err != nil {
		return errors.Wrap(err, "couldn't Flush nat rules")
	}

	// allow masquerading
	if err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", masqueradeIp, "-o", netInterface, "-j", "MASQUERADE").Run(); err != nil {
		return errors.Wrap(err, "couldn't allow masquerading")
	}

	// Allow forwarding between net interface and veth interface
	if err := exec.Command("iptables", "-A", "FORWARD", "-i", netInterface, "-o", vethName, "-j", "ACCEPT").Run(); err != nil {
		return errors.Wrapf(err, "couldn't allow forwarding from net interface %q to veth interface %q", netInterface, vethName)
	}
	if err := exec.Command("iptables", "-A", "FORWARD", "-o", netInterface, "-i", vethName, "-j", "ACCEPT").Run(); err != nil {
		return errors.Wrapf(err, "couldn't allow forwarding from veth interface %q to net interface %q", netInterface, vethName)
	}

	return nil
}

func enableIpv4Forwarding() error {
	if err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run(); err != nil {
		return errors.Wrap(err, "couldn't enable ipv4 forwarding")
	}
	return nil
}

func connectedInterfaceName() (string, error) {
	out, _ := utils.CmdOut("ip", "-4", "route", "ls")

	for _, line := range strings.Split(out, "\n") {
		words := strings.Split(line, " ")
		if words[ipRouteDefaultIndex] == "default" {
			return words[ipRouteNameIndex], nil
		}
	}
	return "", errors.New("Not connected to the internet")
}

// https://play.golang.org/p/BDt3qEQ_2H
// gets local ip
func localIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			var mask net.IPMask
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				mask = v.Mask
			case *net.IPAddr:
				ip = v.IP
				mask = ip.DefaultMask()
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String() + "/" + getIp(mask), nil
		}
	}
	return "", errors.New("couldn't get local IP")
}

// returns Ip given subnet mask
func getIp(mask net.IPMask) string {
	bits := 0
	for i := 0; i < subnetMaskBytes; i++ {
		if mask[i] == subnetLogicOne {
			bits += bitsInByte
		}
	}
	return strconv.Itoa(bits)
}
