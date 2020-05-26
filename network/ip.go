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

type IpAddrManager struct {
	// The path of the JSON format storage file of the allocation status of the subnets.
	SubnetAllocatorPath string
	// Subnet table, each value is bool slice whose elements are indicators of
	// the validity of each IP address in the subnet.
	Subnets map[string][]bool
}

func (ipam *IpAddrManager) Alloc(subnet *net.IPNet) (net.IP, error) {
	subnetNumber := subnet.IP.To4()
	if subnetNumber == nil {
		return nil, fmt.Errorf("Unable to handle non-IPV4 subnet of %v", subnet)
	}

	ipam.Subnets = map[string][]bool{}

	if err := ipam.load(); err != nil {
		log.Warnf("load() error %v", err)
	}

	// If the subnet array is not in the table, initialize it.
	if _, exist := ipam.Subnets[subnet.String()]; !exist {
		// ones is the leading ones in the mask.
		// bits is the total bit length of the mask.
		ones, bits := subnet.Mask.Size()
		if bits-ones >= strconv.IntSize {
			return nil, fmt.Errorf("Subnet range of ones %d and bits %d is too big", ones, bits)
		}
		// Default values are all false.
		ipam.Subnets[subnet.String()] = make([]bool, 1<<uint(bits-ones))
	}

	defer func() {
		go ipam.dump()
	}()

	for i, c := range ipam.Subnets[subnet.String()] {
		// Skip allocated addresses and the first and last addresses,
		// since they are used for representing the subnet and broadcasting.
		if !c && i != 0 && i < len(ipam.Subnets[subnet.String()])-1 {
			ipam.Subnets[subnet.String()][i] = true

			// Perfectly clone a shallow copy of the original byte slice.
			// https://github.com/go101/go101/wiki/How-to-perfectly-clone-a-slice
			ip := append(subnetNumber[:0:0], subnetNumber...).Mask(subnet.Mask)
			for j := 0; j < 4; j++ {
				ip[j] += byte(i >> uint(3-j) * 8)
			}

			return ip, nil
		}
	}

	return nil, fmt.Errorf("No usable remaining IP address in the subnet")
}

func (ipam *IpAddrManager) Release(subnet *net.IPNet, ip net.IP) error {
	ipam.Subnets = map[string][]bool{}

	if err := ipam.load(); err != nil {
		return fmt.Errorf("load() error %v", err)
	}

	subnetNumber := subnet.IP.To4()
	if subnetNumber == nil {
		return fmt.Errorf("Unable to handle non-IPV4 subnet address of %s", subnet.IP)
	}

	ipToRelease := ip.To4()
	if ipToRelease == nil {
		return fmt.Errorf("Unable to handle non-IPV4 subnet address of %s", ip)
	}

	// Perfectly clone a shallow copy of the original byte slice.
	// https://github.com/go101/go101/wiki/How-to-perfectly-clone-a-slice
	subnetIp := subnetNumber.Mask(subnet.Mask)
	idx := 0
	for i := 0; i < 4; i++ {
		idx += int(ipToRelease[i]-subnetIp[i]) << uint((3-i)*8)
	}
	if idx == 0 || idx >= len(ipam.Subnets[subnet.String()])-1 {
		return fmt.Errorf("Invalid allocated IP address %v of subnet %v", ipToRelease, subnet)
	}

	ipam.Subnets[subnet.String()][idx] = false
	defer func() {
		go ipam.dump()
	}()

	return nil
}

// Generally, load() needs an empty subnet table.
func (ipam *IpAddrManager) load() error {
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

func (ipam *IpAddrManager) dump() {
	configFileDir, _ := filepath.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(configFileDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(configFileDir, 0644); err != nil {
				log.Errorf("MkdirAll() %s error %v", configFileDir, err)
			}
		} else {
			log.Errorf("Stat() %s error %v", configFileDir, err)
		}
	}

	configJsonBytes, err := json.Marshal(ipam.Subnets)
	if err != nil {
		log.Errorf("Marshal() error %v", err)
	}

	configFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer configFile.Close()
	if err != nil {
		log.Errorf("OpenFile() %s error %v", ipam.SubnetAllocatorPath, err)
	}

	if _, err := configFile.Write(configJsonBytes); err != nil {
		log.Errorf("Write() to file %s error %v", ipam.SubnetAllocatorPath, err)
	}
}
