package main

import (
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
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

func createNetConnectivity() error {
	nsName := "lockerNs"
	vethName := "v-locker"
	vethPeerName := "v-locker-peer"
	vethIp := "10.200.1.1"
	vethCIDR := vethIp + "/24"
	vethPeerCIDR := "10.200.1.2/24"
	loopback := "lo"
	masqueradeIp := "10.200.1.0/255.255.255.0"
	netInterface, err := connectedInterfaceName()
	if err != nil {
		return err
	}

	// create network namespace
	if err := AddNetNs(nsName); err != nil {
		return err
	}

	// create veth pair
	if err := addVethPair(vethName, vethPeerName); err != nil {
		return err
	}

	// assign peer to namespace
	assignVethToNs(vethPeerName, nsName)

	// setup ipv4 of veth
	addIp(vethCIDR, vethName)
	if err := setInterfaceUp(vethName); err != nil {
		return err
	}

	// setup ipv4 of veth peer
	addIpInsideNs(vethPeerCIDR, vethPeerName, nsName)
	if err := setInterfaceUpInsideNs(vethPeerName, nsName); err != nil {
		return err
	}
	if err := setInterfaceUpInsideNs(loopback, nsName); err != nil {
		return err
	}

	// add default gateway inside namespace
	addDefaultGateway(nsName, vethIp)

	// enable ipv4 forwarding
	if err := enableIpv4Forwarding(); err != nil {
		return err
	}

	// set rules to allow connectivity
	if err := setIptablesRules(masqueradeIp, netInterface, vethName); err != nil {
		return err
	}

	if err := joinNsByName(nsName); err != nil {
		return err
	}

	return nil
}

func joinNsByName(nsName string) error {
	nsHandle, err := getFdFromPath(netnsDirectory + nsName)
	if err != nil {
		return errors.Wrapf(err, "Couldn't get fd of network namespace %q", nsName)
	}
	return setNs(nsHandle, syscall.CLONE_NEWNET)
}

func setNs(nsHandle int, nsType int) error {
	if _, _, err := syscall.Syscall(SYS_SETNS, uintptr(nsHandle), uintptr(nsType), 0); err != 0 {
		return errors.Wrap(err, "Couldn't set network namespace")
	}
	return nil
}

func AddNetNs(nsName string) error {

	if netNsExists(nsName) {
		return nil
	}

	if err := exec.Command("ip", "netns", "add", nsName).Run(); err != nil {
		return errors.Wrapf(err, "Couldn't add new network namespace %q", nsName)
	}
	return nil
}

// function return true if namespace exists
func netNsExists(nsName string) bool {
	out, _ := cmdOut("ip", "netns", "list")
	namespaces := strings.Split(out, "\n")
	return stringInSlice(nsName, namespaces)
}

func addVethPair(vethName, vethPeerName string) error {

	if netInterfaceExists(vethName) {
		return nil
	}

	if err := exec.Command("ip", "link", "add", vethName, "type", "veth", "peer", "name", vethPeerName).Run(); err != nil {
		return errors.Wrapf(err, "Couldn't add veth pair of %q and %q", vethName, vethPeerName)
	}
	return nil
}

// function return true if Veth pair exists
func netInterfaceExists(vethName string) bool {
	out, _ := cmdOut("ip", "link", "list")
	return strings.Contains(out, vethName+"@")
}

func assignVethToNs(vethName, nsName string) {
	cmd := exec.Command("ip", "link", "set", vethName, "netns", nsName)
	cmd.Run()
}

func addIp(Ip, vethName string) {
	cmd := exec.Command("ip", "addr", "add", Ip, "dev", vethName)
	cmd.Run() //if fails, ip already exists
}

func addIpInsideNs(Ip, vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", Ip, "dev", vethName)
	cmd.Run() //if fails, ip already exists
}

func setInterfaceUpInsideNs(vethName, nsName string) error {
	if err := exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethName, "up").Run(); err != nil {
		return errors.Wrapf(err, "Couldn't set interface %q up inside namespace %q", vethName, nsName)
	}
	return nil
}

func bridgeExists(bridgeName string) bool {
	out, _ := cmdOut("ip", "link", "list", "type", "bridge")
	return strings.Contains(out, bridgeName)
}

func createBridge(bridgeName string) error {
	if bridgeExists(bridgeName) {
		return nil
	}

	if err := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge").Run(); err != nil {
		return errors.Wrapf(err, "Couldn't create bridge %q", bridgeName)
	}
	return nil
}

func setInterfaceUp(interfaceName string) error {
	if err := exec.Command("ip", "link", "set", interfaceName, "up").Run(); err != nil {
		return errors.Wrapf(err, "Couldn't set interface %q up", interfaceName)
	}
	return nil
}

func addInterfaceToBridge(interfaceName, bridgeName string) error {
	if err := exec.Command("ip", "link", "set", interfaceName, "master", bridgeName).Run(); err != nil {
		return errors.Wrapf(err, "Couldn't add bridge %q to interface %q", bridgeName, interfaceName)
	}
	return nil
}

func addBridgeIp(bridgeName, Ip string) error {
	if err := exec.Command("ip", "addr", "add", Ip, "brd", "+", "dev", bridgeName).Run(); err != nil {
		return errors.Wrapf(err, "Couldn't add ip %q to bridge %q", Ip, bridgeName)
	}
	return nil
}

func addDefaultGateway(nsName, Ip string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "route", "add", "default", "via", Ip)
	cmd.Run() // if fails, route already exists
}

func setIptablesRules(masqueradeIp, netInterface, vethName string) error {
	// Policy DROP by default.
	if err := exec.Command("iptables", "-P", "FORWARD", "DROP").Run(); err != nil {
		return errors.Wrap(err, "Couldn't policy DROP by default")
	}

	// Flush forward rules,
	if err := exec.Command("iptables", "-F", "FORWARD").Run(); err != nil {
		return errors.Wrap(err, "Couldn't flush forward rules")
	}

	// Flush nat rules.
	if err := exec.Command("iptables", "-t", "nat", "-F").Run(); err != nil {
		return errors.Wrap(err, "Couldn't Flush nat rules")
	}

	// allow masquerading
	if err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", masqueradeIp, "-o", netInterface, "-j", "MASQUERADE").Run(); err != nil {
		return errors.Wrap(err, "Couldn't allow masquerading")
	}

	// Allow forwarding between net interface and veth interface
	if err := exec.Command("iptables", "-A", "FORWARD", "-i", netInterface, "-o", vethName, "-j", "ACCEPT").Run(); err != nil {
		return errors.Wrapf(err, "Couldn't allow forwarding from net interface %q to veth interface %q", netInterface, vethName)
	}
	if err := exec.Command("iptables", "-A", "FORWARD", "-o", netInterface, "-i", vethName, "-j", "ACCEPT").Run(); err != nil {
		return errors.Wrapf(err, "Couldn't allow forwarding from veth interface %q to net interface %q", netInterface, vethName)
	}

	return nil
}

func enableIpv4Forwarding() error {
	if err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run(); err != nil {
		return errors.Wrap(err, "Couldn't enable ipv4 forwarding")
	}
	return nil
}

func connectedInterfaceName() (string, error) {
	out, _ := cmdOut("ip", "-4", "route", "ls")

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
	return "", errors.New("Couldn't get local IP")
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
