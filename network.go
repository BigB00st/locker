package main

import (
	"errors"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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

func createNetConnectivity() {
	nsName := "lockerNs"
	vethName := "v-locker"
	vethPeerName := "v-locker-peer"
	vethIp := "10.200.1.1"
	vethCIDR := vethIp + "/24"
	vethPeerCIDR := "10.200.1.2/24"
	loopback := "lo"
	masqueradeIp := "10.200.1.0/255.255.255.0"
	netInterface, err := connectedInterfaceName()
	must(err)

	// create network namespace
	AddNetNs(nsName)

	// create veth pair
	addVethPair(vethName, vethPeerName)

	// assign peer to namespace
	assignVethToNs(vethPeerName, nsName)

	// setup ipv4 of veth
	addIp(vethCIDR, vethName)
	setInterfaceUp(vethName)

	// setup ipv4 of veth peer
	addIpInsideNs(vethPeerCIDR, vethPeerName, nsName)
	setInterfaceUpInsideNs(vethPeerName, nsName)
	setInterfaceUpInsideNs(loopback, nsName)

	// add default gateway inside namespace
	addDefaultGateway(nsName, vethIp)

	// enable ipv4 forwarding
	enableIpv4Forwarding()

	// set rules to allow connectivity
	setIptablesRules(masqueradeIp, netInterface, vethName)

	joinNsByName(nsName)
}

func joinNsByName(nsName string) {
	nsHandle, err := getFdFromPath(netnsDirectory + nsName)
	must(err)
	must(setNs(nsHandle, syscall.CLONE_NEWNET))
}

func setNs(nsHandle int, nsType int) (err error) {
	_, _, e1 := syscall.Syscall(SYS_SETNS, uintptr(nsHandle), uintptr(nsType), 0)
	if e1 != 0 {
		err = e1
	}
	return
}

func AddNetNs(nsName string) {

	if netNsExists(nsName) {
		return
	}

	cmd := exec.Command("ip", "netns", "add", nsName)
	must(cmd.Run())
}

// function return true if namespace exists
func netNsExists(nsName string) bool {
	out, _ := cmdOut("ip", "netns", "list")
	namespaces := strings.Split(out, "\n")
	return stringInSlice(nsName, namespaces)
}

func addVethPair(vethName, vethPeerName string) {

	if netInterfaceExists(vethName) {
		return
	}

	cmd := exec.Command("ip", "link", "add", vethName, "type", "veth", "peer", "name", vethPeerName)
	must(cmd.Run())
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

func setInterfaceUpInsideNs(vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethName, "up")
	must(cmd.Run())
}

func bridgeExists(bridgeName string) bool {
	out, _ := cmdOut("ip", "link", "list", "type", "bridge")
	return strings.Contains(out, bridgeName)
}

func createBridge(bridgeName string) {
	if bridgeExists(bridgeName) {
		return
	}

	cmd := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge")
	must(cmd.Run())
}

func setInterfaceUp(interfaceName string) {
	cmd := exec.Command("ip", "link", "set", interfaceName, "up")
	must(cmd.Run())
}

func addInterfaceToBridge(interfaceName, bridgeName string) {
	cmd := exec.Command("ip", "link", "set", interfaceName, "master", bridgeName)
	must(cmd.Run())
}

func addBridgeIp(bridgeName, Ip string) {
	cmd := exec.Command("ip", "addr", "add", Ip, "brd", "+", "dev", bridgeName)
	must(cmd.Run())
}

func addDefaultGateway(nsName, Ip string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "route", "add", "default", "via", Ip)
	cmd.Run() // if fails, route already exists
}

func setIptablesRules(masqueradeIp, netInterface, vethName string) {
	// Flush forward rules, policy DROP by default.
	cmd := exec.Command("iptables", "-P", "FORWARD", "DROP")
	must(cmd.Run())
	cmd = exec.Command("iptables", "-F", "FORWARD")
	must(cmd.Run())

	// Flush nat rules.
	cmd = exec.Command("iptables", "-t", "nat", "-F")
	must(cmd.Run())

	// allow masquerading
	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", masqueradeIp, "-o", netInterface, "-j", "MASQUERADE")
	must(cmd.Run())

	// Allow forwarding between net interface and veth interface
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", netInterface, "-o", vethName, "-j", "ACCEPT")
	must(cmd.Run())
	cmd = exec.Command("iptables", "-A", "FORWARD", "-o", netInterface, "-i", vethName, "-j", "ACCEPT")
	must(cmd.Run())
}

func enableIpv4Forwarding() {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	must(cmd.Run())
}

func connectedInterfaceName() (string, error) {
	out, _ := cmdOut("ip", "-4", "route", "ls")

	for _, line := range strings.Split(out, "\n") {
		words := strings.Split(line, " ")
		if words[ipRouteDefaultIndex] == "default" {
			return words[ipRouteNameIndex], nil
		}
	}
	return "", errors.New("Not connected to internet")
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
	return "", errors.New("are you connected to the network?")
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
