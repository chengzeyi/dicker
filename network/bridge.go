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

func (b *BridgeNetworkDriver) CreateNetwork(nwName, subnet string, gatewayIp net.IP) (*Network, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("ParseCIDR() of %s error %v", subnet, err)
	}

	if err := b.createBridgeInterface(nwName); err != nil {
		return nil, fmt.Errorf("createBridgeInterface() %s error %v", nwName, err)
	}

	if err := b.setInterfaceIp(nwName, &net.IPNet{IP: gatewayIp, Mask: ipNet.Mask}); err != nil {
		return nil, fmt.Errorf("setInterfaceIp() of bridge interface %s, gatewayIp %v and subnet mask %v error %v", nwName, gatewayIp, ipNet.Mask, err)
	}

	if err := b.setUpInterface(nwName); err != nil {
		return nil, fmt.Errorf("setUpInterface() of bridge interface %s error %v", nwName, err)
	}

	if err := b.setUpIpTables(nwName, ipNet); err != nil {
		return nil, fmt.Errorf("setUpInterface() of bridge interface %s and subnet %v error %v", nwName, ipNet, err)
	}

	return &Network{
		Name: nwName,
		// The transformed notation can automatically suit the v4 or v6 format.
		Subnet: ipNet.String(),
		GatewayIp: gatewayIp,
		Driver: b.Name(),
	}, nil
}

func (b *BridgeNetworkDriver) DeleteNetwork(nw *Network) error {
	nwName := nw.Name
	iface, err := b.findInterface(nwName)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", nwName, err)
	}

	if err := netlink.LinkDel(iface); err != nil {
		return fmt.Errorf("LinkDel() interface %s error %v", nwName, err)
	}

	return nil
}

// Connect the endpoint to the network.
func (b *BridgeNetworkDriver) ConnectToNetwork(nw *Network, endpoint *Endpoint) error {
	nwName := nw.Name
	iface, err := b.findInterface(nwName)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", nwName, err)
	}

	linkAttrs := netlink.NewLinkAttrs()
	// The name must be less than 16 characters.
	linkAttrs.Name = endpoint.Id[:8]
	linkAttrs.MasterIndex = iface.Attrs().Index

	endpoint.Device = &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  "cif-" + endpoint.Id[:8],
	}

	if err := netlink.LinkAdd(endpoint.Device); err != nil {
		return fmt.Errorf("LinkAdd() veth with peer name %s error %v", endpoint.Device.PeerName, err)
	}

	if err := netlink.LinkSetUp(endpoint.Device); err != nil {
		return fmt.Errorf("LinkSetUp() veth with peer name %s error %v", endpoint.Device.PeerName, err)
	}

	return nil
}

// Disconnect the endpoint from the network.
func (b *BridgeNetworkDriver) DisconnectFromNetwork(nw *Network, endpoint *Endpoint) error {
	panic("not implemented")
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

// Set the Ip of the network interface.
// The IP contains the mask, so it is represented as net.IPNet.
func (b *BridgeNetworkDriver) setInterfaceIp(name string, ip *net.IPNet) error {
	iface, err := b.findInterface(name)
	if err != nil {
		return fmt.Errorf("findInterface() interface %s error %v", name, err)
	}

	addr := &netlink.Addr{
		IPNet: ip,
		Peer:  ip,
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
		// Operate on the nat table.
		"-t", "nat",
		// Append POSTROUTING rule to the end of the selected chain.
		// POSTROUTING: for altering packets as they are about to go out.
		"-A", "POSTROUTING",
		// Specify the source to be in the network.
		"-s", subnet.String(),
		// Do not match packets sent via this interface.
		"!", "-o", name,
		// Specify the target of the rule.
		// MASQUERADE: This target is only valid in the nat table, in the POSTROUTING
		// chain.
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
