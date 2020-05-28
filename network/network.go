package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/chengzeyi/dicker/container"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	DEFAULT_NETWORK_PATH = "/var/run/dicker/network/network"
)

var (
	drivers  = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

func GetNetworkDriver(name string) NetworkDriver {
	return drivers[name]
}

func GetNetwork(name string) *Network {
	return networks[name]
}

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
	CreateNetwork(nwName, subnet string, gatewayIp net.IP) (*Network, error)
	DeleteNetwork(nw *Network) error
	ConnectToNetwork(nw *Network, endpoint *Endpoint) error
	DisconnectFromNetwork(nw *Network, endpoint *Endpoint) error
}

type Network struct {
	Name      string `json:"name"`
	Subnet    string `json:"subnet"`
	GatewayIp net.IP `json:"gateway_ip"`
	Driver    string `json:"driver"`
}

func (n *Network) Dump(path string) error {
	dirPath, _ := filepath.Split(path)
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0644); err != nil {
				return fmt.Errorf("MkdirAll() %s error %v", dirPath, err)
			}
		} else {
			return fmt.Errorf("Stat() %s error %v", dirPath, err)
		}
	}

	// O_TRUNC:: clear the file before writing.
	nwFile, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("OpenFile() %s error %v", path, err)
	}
	defer nwFile.Close()

	nwJsonBytes, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("Marshal() %v error %v", n, err)
	}

	if _, err := nwFile.Write(nwJsonBytes); err != nil {
		return fmt.Errorf("Write() to %s error %v", path, err)
	}

	return nil
}

func (n *Network) Remove(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return fmt.Errorf("Stat() %s error %v", path, err)
		}
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("Remove() %s error %v", path, err)
	}

	return nil
}

func (n *Network) Load(path string) error {
	nwFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Open() %s error %v", path, err)
	}

	nwJsonBytes := make([]byte, 2048)
	numBytes, err := nwFile.Read(nwJsonBytes)
	if err != nil {
		return fmt.Errorf("Read() from %s error %v", path, err)
	}

	if err := json.Unmarshal(nwJsonBytes[:numBytes], n); err != nil {
		return fmt.Errorf("Unmarshal() error %v", err)
	}

	return nil
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

		nw := &Network{}
		if err := nw.Load(nwPath); err != nil {
			log.Errorf("Load() %s error %v", nwPath, err)
		}
		log.Infof("Network %s loaded", nw.Name)

		networks[nw.Name] = nw
		return nil
	})

	return nil
}

// Create a new network in the subnet with the driver.
func CreateNetwork(driver, subnet, nwName string) error {
	netDriver := GetNetworkDriver(driver)
	if netDriver == nil {
		return fmt.Errorf("Invalid network driver name %s", driver)
	}

	gatewayIp, err := ipAllocator.Alloc(subnet)
	if err != nil {
		return fmt.Errorf("Alloc() in net %s error %v", subnet, err)
	}
	
	nw, err := netDriver.CreateNetwork(nwName, subnet, gatewayIp)
	if err != nil {
		return fmt.Errorf("CreateNetwork() %s with subnet %s and gateway IP %v error", nwName, subnet, gatewayIp)
	}

	nwPath := filepath.Join(DEFAULT_IP_ADDR_MANAGER_ALLOCATOR_PATH, nwName)
	if err := nw.Dump(nwPath); err != nil {
		return fmt.Errorf("Dump() %s error %v", nwPath, err)
	}

	networks[nwName] = nw

	return nil
}

func ListNetwork() error {
	writer := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(writer, "NAME\tSUBNET\tGATEWAY_IP\tDRIVER\n")
	for _, nw := range networks {
		fmt.Fprintf(writer, "%s\t%s\t%v\t%s\n", nw.Name, nw.Subnet, nw.GatewayIp, nw.Driver)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Flush() error %v", err)
	}

	return nil
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
