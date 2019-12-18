package main

// https://github.com/teddyking/netsetgo

import (
	"fmt"
	"net"
	"os"
	//"github.com/vishvananda/netlink"
)
/*
route := &netlink.Route{
	Scope:     netlink.SCOPE_UNIVERSE,
	LinkIndex: link.Attrs().Index,
	Gw:        netConfig.BridgeIP,
}

return netlink.RouteAdd(route)*/

func runNetwork() {

	var err error = nil
	
	bridge := NewBridge();
	bridgeName := "bridgename";
	bridgeIP, bridgeSubnet, err := net.ParseCIDR("10.10.10.1/24");
	if err != nil {
		fmt.Println("Error parsing CIDR")
		os.Exit(1)
	}

	// create bridge
	bridgeInterface, err := bridge.Create(bridgeName, bridgeIP, bridgeSubnet)
	if err != nil {
		fmt.Println("Error creating bridge")
		os.Exit(1)
	}

	// create veth pair
	vethNamePrefix := "veth"
	veth := NewVeth()
	hostVeth, containerVeth, err := veth.Create(vethNamePrefix)
	if err != nil {
		fmt.Println("Error creating veth pair")
		os.Exit(1)
	}

	// attach bridge to host veth
	err = bridge.Attach(bridgeInterface, hostVeth)
	if err != nil {
		fmt.Println("Error attaching bridge")
		os.Exit(1)
	}

	// move container veth to new network
	err = veth.MoveToNetworkNamespace(containerVeth, os.Getpid())
	if err != nil {
		fmt.Println("Error moving to network namespace")
		os.Exit(1)
	}


}
