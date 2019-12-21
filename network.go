package main


import (
	"os/exec"
	"bytes"
	"strings"
	"errors"
	"net"
	"strconv"
)

func networkMain(){
	/*localIp, err  := localIP()
	must(err)*/

	localIp := "192.168.1.0/24"

	ns1Name := "namespace1"
	ns2Name := "namespace2"
	veth1Name := "veth1"
	veth2Name := "veth2"
	brVeth1Name := "br1-veth"
	brVeth2Name := "br2-veth"
	veth1CIDR := "192.168.1.11/24"
	veth2CIDR := "192.168.1.12/24"
	bridgeName := "br1"
	bridgeIp := "192.168.1.10"
	bridgeCIDR := bridgeIp + "/24"

	// create network namespaces
	AddNetNs(ns1Name)
	AddNetNs(ns2Name)

	// create veth pairs
	addVethPair(veth1Name, brVeth1Name)
	addVethPair(veth2Name, brVeth2Name)

	// assign veths to namespaces
	assignVethToNs(veth1Name, ns1Name)
	assignVethToNs(veth2Name, ns2Name)

	// add ip to veths inside namespace
	addIpInsideNs(veth1CIDR, veth1Name, ns1Name)
	addIpInsideNs(veth2CIDR, veth2Name, ns2Name)
	
	// create bridge and set it up
	createBridge(bridgeName)
	setInterfaceUp(bridgeName)

	// set bridge veths up
	setInterfaceUp(brVeth1Name)
	setInterfaceUp(brVeth2Name)

	// set veths up inside namespace
	setInterfaceUpInsideNs(veth1Name, ns1Name)
	setInterfaceUpInsideNs(veth2Name, ns2Name)

	// add bridge veths to bridge
	addInterfaceToBridge(brVeth1Name, bridgeName)
	addInterfaceToBridge(brVeth2Name, bridgeName)
	
	// add bridge ip
	addBridgeIp(bridgeName, bridgeCIDR)

	// add default gateway in namespaces
	addDefaultGateway(ns1Name, bridgeIp)
	addDefaultGateway(ns2Name, bridgeIp)

	// set rules to allow connectivity
	setIptablesRules(localIp)

	// enable ipv4 forwarding
	enableIpv4Forwarding()

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
	cmd := exec.Command("ip", "netns", "list")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	namespaces := strings.Split(output.String(), "\n")
	return stringInSlice(nsName, namespaces)
}

func addVethPair(vethName, bridgePairName string) {
	
	if vethPairExists(vethName, bridgePairName) {
		return
	}

	cmd := exec.Command("ip", "link", "add", vethName, "type", "veth", "peer", "name", bridgePairName)
	must(cmd.Run())
} 

// function return true if Veth pair exists
func vethPairExists(vethName, bridgePairName string) bool {
	cmd := exec.Command("ip", "link", "list")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	return strings.Contains(output.String(), vethName + "@" + bridgePairName)
}

func assignVethToNs(vethName, nsName string) {
	cmd := exec.Command("ip", "link", "set", vethName, "netns", nsName)
	cmd.Run()
}

func addIpInsideNs(CIDR, vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", CIDR, "dev", vethName)
	cmd.Run()
}

func setInterfaceUpInsideNs(vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethName, "up")
	must(cmd.Run())
}

func bridgeExists(bridgeName string) bool {
	cmd := exec.Command("ip", "link", "list", "type", "bridge")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	return strings.Contains(output.String(), bridgeName)
}

func createBridge(bridgeName string) {
	if bridgeExists(bridgeName) {
		return
	}

	cmd := exec.Command("ip","link", "add", "name", bridgeName, "type", "bridge")
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

func addBridgeIp(bridgeName, CIDR string) {
	cmd := exec.Command("ip", "addr", "add", CIDR, "brd", "+", "dev", bridgeName)
	must(cmd.Run())
}

func addDefaultGateway(nsName, bridgeIp string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "route", "add", "default", "via", bridgeIp)
	must(cmd.Run())
}

func setIptablesRules(localCIDR string) {
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", localCIDR ,"-j", "MASQUERADE")
	must(cmd.Run())
}

func enableIpv4Forwarding() {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	must(cmd.Run())
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
			return ip.String()+"/"+getCIDR(mask), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// returns CIDR given subnet mask
func getCIDR(mask net.IPMask) string {
	bits := 0
	for i := 0; i < subnetMaskBytes; i++ {
		if mask[i] == subnetLogicOne {
			bits += bitsInByte
		}
	}
	return strconv.Itoa(bits)
}