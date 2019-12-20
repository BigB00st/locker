package main


import (
	"os/exec"
	"bytes"
	//"fmt"
	"strings"
)

func AddNetNs(name string) {
	
	if NetNsExists(name) {
		return
	}

	cmd := exec.Command("ip", "netns", "add", name)
	must(cmd.Run())
} 

// function return true if namespace exists
func NetNsExists(name string) bool {
	cmd := exec.Command("ip", "netns", "list")

	//pipe output
	var output bytes.Buffer
	cmd.Stdout = &output

	cmd.Run()
	namespaces := strings.Split(output.String(), "\n")
	return StringInSlice(name, namespaces)
}

func AddVeth(name, bridgePairName string) {
	
	if VethPairExists(name, bridgePairName) {
		return
	}

	cmd := exec.Command("ip", "netns", "add", name)
	must(cmd.Run())
} 

// function return true if Veth pair exists
func VethPairExists(name, bridgePairName string) bool {
	cmd := exec.Command("ip", "netns", "list")

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

func addIpInNs(ip, vethName, nsName string) {
	cmd := exec.Command("ip", "netns", "exec", nsName, "ip", "addr", "add", ip, "dev", vethName)
	cmd.Run()
}
