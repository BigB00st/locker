package main


import (
	"os/exec"
	"bytes"
	//"fmt"
	"strings"
)

func AddNetNs(name string) {
	
	if netNsExists(name) {
		return
	}

	cmd := exec.Command("ip", "netns", "add", name)
	must(cmd.Run())
} 

// function return true if namespace exists
func netNsExists(name string) bool {
	cmd := exec.Command("ip", "netns", "list")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	namespaces := strings.Split(output.String(), "\n")
	return stringInSlice(name, namespaces)
}

func addVeth(name, bridgePairName string) {
	
	if vethPairExists(name, bridgePairName) {
		return
	}

	cmd := exec.Command("ip", "netns", "add", name)
	must(cmd.Run())
} 

// function return true if Veth pair exists
func vethPairExists(name, bridgePairName string) bool {
	cmd := exec.Command("ip", "link", "list")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	return strings.Contains(output.String(), name + "@" + bridgePairName)
}

func assignVethToNs(vethName, nsName string) {
	cmd := exec.Command("ip", "link", "set", vethName, "netns", nsName)
	cmd.Run()
}

func addIpInsideNs(ip, vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", ip, "dev", vethName)
	cmd.Run()
}

func setInterfaceUpInsideNs(vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "link", "set", vethName, "up")
	must(cmd.Run())
}

func bridgeExists(name string) bool {
	cmd := exec.Command("ip", "link", "list", "type", "bridge")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	return strings.Contains(output.String(), name)
}

func createBridge(name string) {
	if bridgeExists(name) {
		return
	}

	cmd := exec.Command("ip","link", "add", "name", name, "type", "bridge")
	must(cmd.Run())
}

func setInterfaceUp(name string) {
	cmd := exec.Command("ip", "link", "set", name, "up")
	must(cmd.Run())
}

func addInterfaceToBridge(name, bridgeName string) {
	cmd := exec.Command("ip", "link", "set", name, "master", bridgeName)
	must(cmd.Run())
}

func configureBridgeInterface(name, ip string) {
	cmd := exec.Command("ip", "addr", "add", ip, "brd", "+", "dev", name)
	must(cmd.Run())
}

func setIptablesRules(nsIp string) {
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", nsIp,"-j", "MASQUERADE")
	must(cmd.Run())
}

func enableIpv4Forwarding() {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	must(cmd.Run())
}
