package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	DEFAULT_IP_ADDR_MANAGER_ALLOCATOR_PATH = "/var/run/dicker/network/ipam/subnet.json"
)

type IpAddrManager struct {
	// The path of the JSON format storage file of the allocation status of the subnets.
	SubnetAllocatorPath string
	// Subnet table, each value is bool slice whose elements are indicators of
	// the validity of each IP address in the subnet.
	Subnets map[string][]bool
}

var ipAllocator = &IpAddrManager{
	SubnetAllocatorPath: DEFAULT_IP_ADDR_MANAGER_ALLOCATOR_PATH,
}

func (ipam *IpAddrManager) Alloc(subnet string) (net.IP, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("ParseCIDR() %s error %v", subnet, err)
	}

	subnetNumber := ipNet.IP.To4()
	if subnetNumber == nil {
		return nil, fmt.Errorf("Unable to handle non-IPV4 IP net of %v", ipNet)
	}

	if err := ipam.load(); err != nil {
		log.Warnf("load() error %v", err)
	}

	// If the subnet array is not in the table, initialize it.
	if _, exist := ipam.Subnets[ipNet.String()]; !exist {
		// ones is the leading ones in the mask.
		// bits is the total bit length of the mask.
		ones, bits := ipNet.Mask.Size()
		if bits-ones >= strconv.IntSize {
			return nil, fmt.Errorf("Subnet range of ones %d and bits %d is too big", ones, bits)
		}
		// Default values are all false.
		ipam.Subnets[ipNet.String()] = make([]bool, 1<<uint(bits-ones))
	}

	for i, c := range ipam.Subnets[ipNet.String()] {
		// Skip allocated addresses and the first and last addresses,
		// since they are used for representing the subnet and broadcasting.
		if !c && i != 0 && i < len(ipam.Subnets[ipNet.String()])-1 {
			// Perfectly clone a shallow copy of the original byte slice.
			// https://github.com/go101/go101/wiki/How-to-perfectly-clone-a-slice
			ip := append(subnetNumber[:0:0], subnetNumber...)
			for j := 0; j < 4; j++ {
				ip[j] += byte(i >> (uint(3-j) * 8))
			}

			ip16 := ip.To16()
			if ip16 == nil {
				return nil, fmt.Errorf("Cannot convert IP %v to a 16-bit representation", ip)
			}

			ipam.Subnets[ipNet.String()][i] = true
			if err := ipam.dump(); err != nil {
				return nil, fmt.Errorf("dump() error %v", err)
			}

			return ip16, nil
		}
	}

	return nil, fmt.Errorf("No usable remaining IP address in the IP net %s", subnet)
}

func (ipam *IpAddrManager) Release(subnet string, ip net.IP) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("ParseCIDR() %s error %v", subnet, err)
	}

	subnetNumber := ipNet.IP.To4()
	if subnetNumber == nil {
		return fmt.Errorf("Unable to handle non-IPV4 IP net %v", ipNet)
	}

	ipToRelease := ip.To4()
	if ipToRelease == nil {
		return fmt.Errorf("Unable to handle non-IPV4 IP %v", ip)
	}

	if err := ipam.load(); err != nil {
		return fmt.Errorf("load() error %v", err)
	}

	idx := 0
	for i := 0; i < 4; i++ {
		idx += int(ipToRelease[i]-subnetNumber[i]) << uint((3-i)*8)
	}
	if idx == 0 || idx >= len(ipam.Subnets[ipNet.String()])-1 {
		return fmt.Errorf("Invalid allocated IP address %v of IP net %v", ipToRelease, subnet)
	}

	ipam.Subnets[ipNet.String()][idx] = false

	if err := ipam.dump(); err != nil {
		return fmt.Errorf("dump() error %v", err)
	}

	return nil
}

func (ipam *IpAddrManager) load() error {
	ipam.Subnets = map[string][]bool{}

	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Path %s not exists", ipam.SubnetAllocatorPath)
		}
		return fmt.Errorf("Stat() %s error %v", ipam.SubnetAllocatorPath, err)
	}

	configFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer configFile.Close()
	if err != nil {
		return fmt.Errorf("Open() %s error %v", ipam.SubnetAllocatorPath, err)
	}

	subnetJsonBytes := make([]byte, 2048)
	n, err := configFile.Read(subnetJsonBytes)
	if err != nil {
		return fmt.Errorf("Read() from file %s error %v", ipam.SubnetAllocatorPath, err)
	}

	if err := json.Unmarshal(subnetJsonBytes[:n], &ipam.Subnets); err != nil {
		return fmt.Errorf("Unmarshal() error %v", err)
	}

	return nil
}

func (ipam *IpAddrManager) dump() error {
	configFileDir, _ := filepath.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(configFileDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(configFileDir, 0644); err != nil {
				return fmt.Errorf("MkdirAll() %s error %v", configFileDir, err)
			}
		} else {
			return fmt.Errorf("Stat() %s error %v", configFileDir, err)
		}
	}

	configJsonBytes, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return fmt.Errorf("Marshal() %v error %v", ipam.Subnets, err)
	}

	configFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer configFile.Close()
	if err != nil {
		return fmt.Errorf("OpenFile() %s error %v", ipam.SubnetAllocatorPath, err)
	}

	if _, err := configFile.Write(configJsonBytes); err != nil {
		return fmt.Errorf("Write() to file %s error %v", ipam.SubnetAllocatorPath, err)
	}

	return nil
}
