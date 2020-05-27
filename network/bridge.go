package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) CreateNetwork(subnet string, name string) (*Network, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("ParseCIDR() of %s error %v", subnet, err)
	}

	net := &Network{
		Name:   name,
		IpNet:  ipNet,
		Driver: b.Name(),
	}

	if err := b.initBridge(net); err != nil {
		return nil, fmt.Errorf("initBridge() of subnet %s error %v", subnet, err)
	}

	return net, nil
}

func (b *BridgeNetworkDriver) DeleteNetwork(nw *Network) error {
	ifaceName  := nw.Name
	iface, err := b.findInterface(ifaceName)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", ifaceName, err)
	}

	if err := netlink.LinkDel(iface); err != nil {
		return fmt.Errorf("LinkDel() interface %s error %v", ifaceName, err)
	}

	return nil
}

func (b *BridgeNetworkDriver) ConnectToNetwork(nw *Network, endpoint *Endpoint) error {
	ifaceName := nw.Name
	iface, err := b.findInterface(ifaceName)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", ifaceName, err)
	}
	
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = endpoint.Id
	linkAttrs.MasterIndex = iface.Attrs().Index

	endpoint.Device = &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName: "cif-" + endpoint.Id,
	}

	panic("not implemented")
}

func (b *BridgeNetworkDriver) DisconnectFromNetwork(nw *Network, endpoint *Endpoint) error {
	panic("not implemented") // TODO: Implement
}

func (b *BridgeNetworkDriver) initBridge(net *Network) error {
	bridgeName := net.Name

	if err := b.createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("createBridgeInterface() %s error %v", bridgeName, err)
	}

	if err := b.setInterfaceIpNet(bridgeName, net.IpNet); err != nil {
		return fmt.Errorf("setInterfaceIpNet() of bridge interface %s and IP net %v error %v", bridgeName, net.IpNet, err)
	}

	if err := b.setUpInterface(bridgeName); err != nil {
		return fmt.Errorf("setUpInterface() of bridge interface %s error %v", bridgeName, err)
	}
	
	if err := b.setUpIpTables(bridgeName, net.IpNet); err != nil {
		return fmt.Errorf("setUpInterface() of bridge interface %s and IP net error %v", bridgeName, net.IpNet, err)
	}
	
	return nil
}

// Create a new network bridge interface with bridgeName as its name.
func (b *BridgeNetworkDriver) createBridgeInterface(name string) error {
	_, err := net.InterfaceByName(name)
	if err == nil {
		return fmt.Errorf("Interface %s already exists", name)
	}
	opErr, ok := err.(*net.OpError)
	if !ok {
		return fmt.Errorf("Check existence of interface %s error %v", name, err)
	}
	_, ok = opErr.Unwrap().(*os.SyscallError)
	if ok {
		return fmt.Errorf("Check existence of interface %s error %v", name, err)
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = name
	bridge := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}
	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("LinkAdd() of interface %s error %v", name, err)
	}

	return nil
}

// Set the Ip net of the network interface.
func (b *BridgeNetworkDriver) setInterfaceIpNet(name string, ipNet *net.IPNet) error {
	iface, err := b.findInterface(name)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", name, err)
	}

	addr := &netlink.Addr{
		IPNet: ipNet,
		Peer:  ipNet,
	}

	if err := netlink.AddrAdd(iface, addr); err != nil {
		return fmt.Errorf("AddrAdd() error %v", err)
	}

	return nil
}

func (b *BridgeNetworkDriver) setUpInterface(name string) error {
	iface, err := b.findInterface(name)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("LinkSetUp() interface of name %s error %v", name, err)
	}

	return nil
}

func (b *BridgeNetworkDriver) setUpIpTables(name string, subnet *net.IPNet) error {
	cmd := exec.Command(
		"iptables",
		"-t", "nat",
		"-A", "POSTROUTING",
		"-s", subnet.String(),
		"!",
		"-o", name,
		"-j", "MASQUERADE",
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Output() command %v with output %v error %v", cmd, output, err)
	}

	return nil
}

func (b *BridgeNetworkDriver) findInterface(name string) (netlink.Link, error) {
	var iface netlink.Link
	var err error
	for i := 0; i < 2; i++ {
		if iface, err = netlink.LinkByName(name); err == nil {
			break
		}
		log.Warnf("LinkByName() of interface %s error %v", name, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("Find network interface %s error %v", name, err)
	}

	return iface, nil
}
