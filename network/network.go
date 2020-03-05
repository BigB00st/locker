package network

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/milosgajdos/tenus"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/vishvananda/netns"
	"gitlab.com/bigboost/locker/utils"
)

const (
	subnetMaskBytes     = 4
	subnetLogicOne      = 255
	bitsInByte          = 8
	ipRouteDefaultIndex = 0
	ipRouteNameIndex    = 4
	netnsDirectory      = "/var/run/netns/"
	ifLen               = 8
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

func CreateConnectivity2() error {
	uuid := viper.GetString("uuid")[:ifLen]
	vethCIDR := "10.200.1.1/24"
	vethPeerCIDR := "10.200.1.2/24"
	vethName := uuid + "0"
	vethPeerName := uuid + "1"
	masqueradeIp := "10.200.1.0/255.255.255.0"
	netInterface, err := connectedInterfaceName()
	if err != nil {
		return err
	}

	veth, err := tenus.NewVethPairWithOptions(vethName, tenus.VethOptions{PeerName: vethPeerName})
	if err != nil {
		return errors.Wrap(err, "couldn't create veth pair")
	}
	if err := veth.SetLinkUp(); err != nil {
		return errors.Wrap(err, "couldn't set veth up")
	}

	// add ip to veth
	vethHostIp, vethHostIpNet, err := net.ParseCIDR(vethCIDR)
	if err != nil {
		return errors.Wrapf(err, "couldn't parse ICDR of %q", vethCIDR)
	}
	if err := veth.SetLinkIp(vethHostIp, vethHostIpNet); err != nil {
		return errors.Wrap(err, "couldn't set ip of veth")
	}

	// enable ipv4 forwarding
	if err := enableIpv4Forwarding(); err != nil {
		return err
	}

	// set rules to allow connectivity
	if err := setIptablesRules(masqueradeIp, netInterface, vethName); err != nil {
		return err
	}

	// get current namespace
	oldNs, err := netns.Get()
	if err != nil {
		return errors.Wrap(err, "couldn't get current network namespace")
	}
	viper.Set("old-netns", oldNs)
	// create new namespace and enter
	newNs, err := netns.New()
	if err != nil {
		return errors.Wrap(err, "couldn't create network namespace")
	}
	// return to old namespace
	if err := netns.Set(oldNs); err != nil {
		return errors.Wrap(err, "couldn't return to old network namespace")
	}

	// assign peer to namespace
	if err := veth.SetPeerLinkNsPid(int(newNs)); err != nil {
		return errors.Wrap(err, "couldn't set veth peer in new network namespace")
	}

	if err := veth.SetPeerLinkUp(); err != nil {
		return errors.Wrap(err, "couldn't set veth peer up")
	}

	// setup ip of veth peer
	vethPeerHostIp, vethPeerHostIpNet, err := net.ParseCIDR(vethPeerCIDR)
	if err != nil {
		return errors.Wrapf(err, "couldn't parse ICDR of %q", vethCIDR)
	}
	if err := veth.SetPeerLinkIp(vethPeerHostIp, vethPeerHostIpNet); err != nil {
		return errors.Wrap(err, "couldn't set ip of veth peer")
	}

	// add default gateway inside namespace
	if err := veth.SetLinkDefaultGw(&vethHostIp); err != nil {
		return errors.Wrap(err, "couldn't set veth peer default gateway")
	}

	// join new namespace
	if err := netns.Set(newNs); err != nil {
		return errors.Wrap(err, "couldn't change to new network namespace")
	}

	return nil
}

func Cleanup() error {
	newNs, _ := netns.Get()
	if err := netns.Set(netns.NsHandle(viper.Get("old-netns"))); err != nil {
		return errors.Wrap(err, "couldn't return to old network namespace")
	}
	vethName := viper.GetString("uuid")[:ifLen] + "0"
	if err := tenus.DeleteLink(vethName); err != nil {
		return errors.Wrapf(err, "couldnt delete link %q", vethName)
	}
	if err := newNs.Close(); err != nil {
		return errors.Wrapf(err, "couldn't close the new network namespace")
	}
	return nil
}

func createVethPair(vethName, vethPeerName, vethCIDR string) (tenus.Linker, error) {
	// create veth pair
	veth, err := tenus.NewVethPairWithOptions(vethName, tenus.VethOptions{PeerName: vethPeerName})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create veth pair")
	}

	// ASSIGN IP ADDRESS TO THE HOST VETH INTERFACE
	vethHostIp, vethHostIpNet, err := net.ParseCIDR(vethCIDR)
	if err != nil {
		log.Fatal(err)
	}

	if err := veth.SetLinkIp(vethHostIp, vethHostIpNet); err != nil {
		fmt.Println(err)
	}

	return veth, nil

}

func CreateConnectivity() error {
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
	if err := addNetNs(nsName); err != nil {
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
		return errors.Wrapf(err, "couldn't set interface %q up inside namespace %q", vethName, nsName)
	}
	return nil
}

func bridgeExists(bridgeName string) bool {
	out, _ := utils.CmdOut("ip", "link", "list", "type", "bridge")
	return strings.Contains(out, bridgeName)
}

/*func createBridge(bridgeName string) error {
	if bridgeExists(bridgeName) {
		return nil
	}

	if err := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge").Run(); err != nil {
		return errors.Wrapf(err, "couldn't create bridge %q", bridgeName)
	}
	return nil
}*/

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

func addDefaultGateway(nsName, Ip string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "route", "add", "default", "via", Ip)
	cmd.Run() // if fails, route already exists
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
