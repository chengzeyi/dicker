package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewWorkspace(parentVolume, containerVolume, imageName, containerName string) error {
	// CreateReadOnlyLayer(imageName)
	// CreateWriteLayer(containerName)
	// CreateMountPoint(containerName, imageName)
	if len(strings.TrimSpace(parentVolume)) == 0 || len(strings.TrimSpace(containerVolume)) == 0 {
		return fmt.Errorf("Invalid argument parentVolume %s or containerVolume %s", parentVolume, containerVolume)
	}

	// MountVolume(volumePaths, containerName)
	// log.Infof("MountVolume new workspace volumes %s", volumePaths)

	return nil
}

func MountVolume(parentVolume, containerVolume, containerName string) error {
	// Create the host directory.
	if err := os.MkdirAll(parentVolume, 0777); err != nil {
		return fmt.Errorf("Mkdir parentVolume %s error %v", err)
	}

	mntPath := filepath.Join(MNT_DIR_PATH, containerName)
	// /root/mnt/containerName/containerVolume is the mount point.
	// It is in the container filesystem.
	containerVolumePath := filepath.Join(mntPath, containerVolume)
	if err := os.MkdirAll(containerVolumePath, 0777); err != nil {
		return fmt.Errorf("Mkdir containerVolumePath %s error %v", err)
	}

	// if err := syscall.Mount() {

	// }
	return nil
}
