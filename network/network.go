package network

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/chengzeyi/dicker/container"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	DEFAULT_NETWORK_PATH = "/var/run/dicker/network/network"
)

var (
	drivers  = map[string]*NetworkDriver{}
	networks = map[string]*Network{}
)

type Endpoint struct {
	Id           string           `json:"id"`
	Device       *netlink.Veth    `json:"device"`
	Ip           net.IP           `json:"ip"`
	Mac          net.HardwareAddr `json:"mac"`
	Network      *Network         `json:"network"`
	PortMappints []string         `json:"port_mappings"`
}

type NetworkDriver interface {
	Name() string
	CreateNetwork(subnet, name string) (*Network, error)
	DeleteNetwork(nw *Network) error
	ConnectToNetwork(nw *Network, endpoint *Endpoint) error
	DisconnectFromNetwork(nw *Network, endpoint *Endpoint) error
}

type Network struct {
	Name   string     `json:"name"`
	IpNet  *net.IPNet `json:"ip_net"`
	Driver string     `json:"driver"`
}

func NewNetWork(name string) *Network {
	return &Network{
		Name: name,
	}
}

func (n *Network) Dump(path string) error {
	panic("not implemented")
}

func (n *Network) Remove(path string) error {
	panic("not implemented")
}

func (n *Network) Load(path string) error {
	panic("not implemented")
}

func Init() error {
	// TODO: init network drivers.

	if _, err := os.Stat(DEFAULT_NETWORK_PATH); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(DEFAULT_NETWORK_PATH, 0644); err != nil {
				return fmt.Errorf("MkdirAll() %s error %v", DEFAULT_NETWORK_PATH, err)
			}
		} else {
			return fmt.Errorf("Stat() %s error %v", DEFAULT_NETWORK_PATH, err)
		}
	}

	filepath.Walk(DEFAULT_NETWORK_PATH, func(nwPath string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			return nil
		}

		_, nwName := filepath.Split(nwPath)
		nw := NewNetWork(nwName)
		if err := nw.Load(nwPath); err != nil {
			log.Errorf("Load() %s error %v", nwPath, err)
		}
		log.Infof("Network %s loaded", nwName)

		networks[nwName] = nw
		return nil
	})

	return nil
}

// Create a new network with the driver and the CIDR format subnet.
func CreateNetwork(driver, subnet, name string) error {
	// _, ipNet, err := net.ParseCIDR(subnet)
	// if err != nil {
	// 	return fmt.Errorf("ParseCIDR() %s error %v", subnet, err)
	// }
	panic("not implemented")
}

func ListNetwork() error {
	panic("not implemented")
}

func DeleteNetwork(name string) error {
	panic("not implemented")
}

func ConnectToNetwork(networkName string, containerInfo *container.ContainerInfo) error {
	panic("not implemented")
}

func DisconnectFromNetwork(networkName string, containerInfo *container.ContainerInfo) error {
	panic("not implemented")
}
