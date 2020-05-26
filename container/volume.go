package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewWorkspace(volumeMapping , imageName, containerName string) error {
	if err := createReadOnlyLayer(imageName); err != nil {
		return fmt.Errorf("createReadOnlyLayer() %s error %v", imageName, err)
	}
	if err := createWriteLayer(containerName); err != nil {
		return fmt.Errorf("createWriteLayer() %s error %v", containerName, err)
	}
	if err := createMountPoint(containerName, imageName); err != nil {
		return fmt.Errorf("createMountPoint() %s with %s error %v", containerName, imageName, err)
	}

	parentVolume, containerVolume, err := parseVolumeMapping(volumeMapping)
	if err != nil {
		log.Errorf("parseVolumeMapping() %s error %v", volumeMapping, err)
	} else if len(parentVolume) != 0 && len(containerVolume) != 0 {
		if err := mountVolume(parentVolume, containerVolume, containerName); err != nil {
			log.Errorf("mountVolume() %s to %s error %v", parentVolume, containerVolume, containerName)
		}
	}

	return nil
}

func parseVolumeMapping(volumeMapping string) (string, string, error) {
	if len(strings.TrimSpace(volumeMapping)) == 0 {
		return "", "", nil
	}
	volumes := strings.Split(volumeMapping, string(os.PathListSeparator))
	if len(volumes) != 2 || len(volumes[0]) == 0 || len(volumes[1]) == 0 {
		return "", "", fmt.Errorf("Invalid volume mapping %s", volumeMapping)
	}
	return volumes[0], volumes[1], nil
}

// Untar image to ROOT_DIR_PATH/imageName
func createReadOnlyLayer(imageName string) error {
	imagePath := filepath.Join(IMAGE_DIR_PATH, imageName+".tar")
	if _, err := os.Stat(imagePath); err != nil {
		return fmt.Errorf("Stat() %s error %v", imagePath, err)	}

	untarFoldPath := filepath.Join(READONLY_LAYER_DIR_PATH, imageName)
	if _, err := os.Stat(untarFoldPath); err != nil {
		if os.IsNotExist(err) {
			// The directory needs creating.
			// log.Infof("Directory %s needs creating", untarFoldPath)
			if err := os.MkdirAll(untarFoldPath, 0622); err != nil {
				return fmt.Errorf("MkdirAll() %s error %v", untarFoldPath, err)
			}
			// After creating the directory, untar the image.
			if _, err := exec.Command("tar", "-xvf", imagePath, "-C", untarFoldPath).CombinedOutput(); err != nil {
				return fmt.Errorf("Untar %s to directory %s error %v", imagePath, untarFoldPath, err)
			}
		} else {
			return fmt.Errorf("Stat() %s exists error %v", untarFoldPath, err)
		}
	}

	// Already exists. No operation needed.
	return nil
}

// Create directory WRITE_LAYER_DIR_PATH/containerName
func createWriteLayer(containerName string) error {
	writePath := filepath.Join(WRITE_LAYER_DIR_PATH, containerName)
	if err := os.MkdirAll(writePath, 0777); err != nil {
		return fmt.Errorf("MkdirAll() %s error %v", writePath, err)
	}

	return nil
}

func createMountPoint(containerName, imageName string) error {
	mntPath := filepath.Join(MNT_DIR_PATH, containerName)
	if err := os.MkdirAll(mntPath, 0777); err != nil {
		return fmt.Errorf("MkdirAll() %s error %v", mntPath, err)
	}

	workDirPath := filepath.Join(OVERLAY_WORK_DIR_PATH, containerName)
	// The Overlay work directory needs to be empty!
	if err := os.MkdirAll(workDirPath, 0777); err != nil {
		return fmt.Errorf("MkdirAll() %s error %v", workDirPath, err)
	}

	writeLayerPath := filepath.Join(WRITE_LAYER_DIR_PATH, containerName)
	readOnlyLayerPath := filepath.Join(READONLY_LAYER_DIR_PATH, imageName)

	options := fmt.Sprintf("upperdir=%s,lowerdir=%s,workdir=%s", writeLayerPath, readOnlyLayerPath, workDirPath)

	if err := syscall.Mount(mntPath, mntPath, "overlay", 0, options); err != nil {
		return fmt.Errorf("Mount() overlay filesystem to %s with options %s error %v", mntPath, options, err)
	}

	return nil
}

// Mount parentVolume to MNT_DIR_PATH/containerName/containerVolume.
func mountVolume(parentVolume, containerVolume, containerName string) error {
	// Create the host directory.
	if err := os.MkdirAll(parentVolume, 0777); err != nil {
		return fmt.Errorf("Mkdir() %s error %v", parentVolume, err)
	}

	mntPath := filepath.Join(MNT_DIR_PATH, containerName)
	// /root/mnt/containerName/containerVolume is the mount point.
	// It is in the container filesystem.
	containerVolumePath := filepath.Join(mntPath, containerVolume)
	if err := os.MkdirAll(containerVolumePath, 0777); err != nil {
		return fmt.Errorf("Mkdir() %s error %v", containerVolumePath, err)
	}

	// Bind parentVolume to containerVolumePath.
	if err := syscall.Mount(parentVolume, containerVolumePath, "", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount() %s to %s error %v", parentVolume, containerVolumePath, err)
	}
	return nil
}

// Delete containerVolume, container mount point and additional layers.
// Return non-nil if any practical deletion operation fails
// and the return value is the last occurred error.
func DeleteWorkspace(volumeMapping, containerName string) error {
	var retErr error
	parentVolume, containerVolume, err := parseVolumeMapping(volumeMapping)
	if err != nil {
		log.Errorf("parseVolumeMapping() %s error %v", volumeMapping, err)
		// Just continue. But if there is really a volume mounted,
		// the unmounting operation of the container mount point will fail.
	} else if len(parentVolume) != 0 && len(containerVolume) != 0 {
		if err := deleteVolume(containerVolume, containerName); err != nil {
			retErr = fmt.Errorf("deleteVolume() %s from %s error %v.", containerVolume, containerName, err)
			log.Error(retErr.Error())
		}
	}
	
	if err := deleteMountPoint(containerName); err != nil {
		retErr = fmt.Errorf("deleteMountPoint() %s error %v", containerName, err)
		log.Error(retErr.Error())
	}
	if err := deleteWriteLayer(containerName); err != nil {
		retErr = fmt.Errorf("deleteWriteLayer() %s error %v", containerName, err)
		log.Error(retErr.Error())
	}

	return retErr
}

// Delete containerVolume from containerName.
// This must be done before deleting the container mount point.
// Or the system will warn the target is busy.
func deleteVolume(containerVolume, containerName string) error {
	mntPath := filepath.Join(MNT_DIR_PATH, containerName)
	containerVolumePath := filepath.Join(mntPath, containerVolume)
	if err := syscall.Unmount(containerVolumePath, 0); err != nil {
		return fmt.Errorf("Unmount() %s error %v", containerVolumePath, err)
	}

	return nil
}

func deleteMountPoint(containerName string) error {
	mntPath := filepath.Join(MNT_DIR_PATH, containerName)
	if err := syscall.Unmount(mntPath, 0); err != nil {
		return fmt.Errorf("Unmount() %s error %v", mntPath, err)
	}
	if err := os.RemoveAll(mntPath); err != nil {
		return fmt.Errorf("RemoveAll() %s error %v", mntPath, err)
	}

	workDirPath := filepath.Join(OVERLAY_WORK_DIR_PATH, containerName)
	// The Overlay work directory needs to be empty!
	if err := os.RemoveAll(workDirPath); err != nil {
		return fmt.Errorf("RemoveAll() %s error %v", workDirPath, err)
	}

	return nil
}

func deleteWriteLayer(containerName string) error {
	writeLayerPath := filepath.Join(WRITE_LAYER_DIR_PATH, containerName)
	if err := os.RemoveAll(writeLayerPath); err != nil {
		return fmt.Errorf("RemoveAll() %s error %v", writeLayerPath, err)
	}

	return nil
}
