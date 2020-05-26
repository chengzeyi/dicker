package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chengzeyi/dicker/container"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenRandStrBytes(n int) string {
	const letterBytes = "1234567890"
	bSlice := make([]byte, n)
	for i := range bSlice {
		bSlice[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(bSlice)
}

func GetContainerPidByName(containerName string) (int, error) {
	containerInfoPath := filepath.Join(container.DEFAULT_INFO_DIR_PATH, containerName)
	configPath := filepath.Join(containerInfoPath, container.CONFIG_FILE_NAME)
	contentBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return 0, err
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return 0, fmt.Errorf("Unmarshal() error %v", err)
	}

	return containerInfo.Pid, nil
}

// How to standardize this?
func GetEnvsByPid(pid string) []string {
	environPath := filepath.Join("/proc", pid, "environ")
	contentBytes, err := ioutil.ReadFile(environPath)
	if err != nil {
		log.Errorf("ReadFile() %s error %v", environPath, err)
		return nil
	}
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}

// Find the root mount point of the cgroup subsystem.
func FindCgroupMountPoint(subsystem string) string {
	// This file contains information about mount points in the process's mount
	// namespace. It supplies various information.
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		log.Errorf("Open() /proc/self/mountinfo error %v", err)
		return ""
	}
	defer f.Close()

	// Baseic format is:
	// 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	// (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Split(text, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				// Mount point: the pathname of the mount point relative to the
				// process's root directory.
				return fields[4]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("Parse /proc/self/mountinfo error %v", err)
	}

	return ""
}

func GetCgroupPath(subsystem, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	cgroupDirPath := filepath.Join(cgroupRoot, cgroupPath)
	if _, err := os.Stat(cgroupDirPath); err != nil {
		if os.IsNotExist(err) && autoCreate {
			if err := os.Mkdir(cgroupDirPath, 0755); err != nil {
				return "", fmt.Errorf("Mkdir() %s error %v", cgroupDirPath, err)
			}
		} else {
			return "", fmt.Errorf("Stat() %s error %v", cgroupDirPath, err)
		}
	}

	return cgroupDirPath, nil
}
