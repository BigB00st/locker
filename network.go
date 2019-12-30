package main


import (
	"os/exec"
	"bytes"
	"strings"
	"errors"
	"net"
	"strconv"
	"fmt"
)

func networkMain(){
	/*localIp, err  := localIP()
	must(err)*/

	nsName := "lockerNs"
	vethName := "v-locker"
	vethPeerName := "v-locker-peer"
	vethIp := "10.200.1.1"
	vethCIDR := vethIp + "/24"
	vethPeerCIDR := "10.200.1.2/24"
	loopback := "lo"
	masqueradeIp := "10.200.1.0/255.255.255.0"
	netInterface := "wlp3s0"

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

func addVethPair(vethName, vethPeerName string) {
	
	if netInterfaceExists(vethName) {
		return
	}

	cmd := exec.Command("ip", "link", "add", vethName, "type", "veth", "peer", "name", vethPeerName)
	must(cmd.Run())
} 

// function return true if Veth pair exists
func netInterfaceExists(vethName string) bool {
	cmd := exec.Command("ip", "link", "list")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	return strings.Contains(output.String(), vethName + "@")
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
	fmt.Println(1)
	cmd = exec.Command("iptables", "-F", "FORWARD")
	must(cmd.Run())
	fmt.Println(2)

	// Flush nat rules.
	cmd = exec.Command("iptables", "-t", "nat", "-F")
	must(cmd.Run())
	fmt.Println(3)

	// allow masquerading
	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING" ,"-s", masqueradeIp, "-o", netInterface, "-j" ,"MASQUERADE")
	must(cmd.Run())
	fmt.Println(4)

	// Allow forwarding between net interface and veth interface
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", netInterface, "-o", vethName, "-j", "ACCEPT")
	must(cmd.Run())
	fmt.Println(5)
	cmd = exec.Command("iptables", "-A", "FORWARD", "-o", netInterface, "-i", vethName, "-j" , "ACCEPT")
	must(cmd.Run())
	fmt.Println(6)
	
	
}

func enableIpv4Forwarding() {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0")
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
			return ip.String()+"/"+getIp(mask), nil
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