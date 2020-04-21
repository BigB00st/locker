package network

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/pkg/io"
)

// setIptablesRules sets rules to allow container connectivity
func setIptablesRules(masqueradeIp, netInterface, vethName string) error {
	// Policy DROP by default.
	if err := exec.Command("iptables", "-P", "FORWARD", "DROP").Run(); err != nil {
		return errors.Wrap(err, "couldn't policy DROP by default")
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

// clearIptablesRules clear iptables rules
func clearIptablesRules(masqueradeIp, netInterface, vethName string) error {
	// clear masquerading
	if str, err := io.CmdOut("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", masqueradeIp, "-o", netInterface, "-j", "MASQUERADE"); err != nil {
		fmt.Println(str)
		return errors.Wrap(err, "couldn't remove masquerading")
	}

	// clear forwarding between net interface and veth interface
	if err := exec.Command("iptables", "-D", "FORWARD", "-i", netInterface, "-o", vethName, "-j", "ACCEPT").Run(); err != nil {
		return errors.Wrapf(err, "couldn't allow forwarding from net interface %q to veth interface %q", netInterface, vethName)
	}
	if err := exec.Command("iptables", "-D", "FORWARD", "-o", netInterface, "-i", vethName, "-j", "ACCEPT").Run(); err != nil {
		return errors.Wrapf(err, "couldn't allow forwarding from veth interface %q to net interface %q", netInterface, vethName)
	}

	return nil
}
