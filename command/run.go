package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chengzeyi/dicker/container"
	"github.com/chengzeyi/dicker/util"

	log "github.com/sirupsen/logrus"
)

type RunOption struct {
	Tty           bool
	ContainerName string
	VolumeMapping string
	PortMappings  []string
	Envs          []string
}

func Run(option *RunOption, imageName string, cmdArr []string) error {
	containerName := option.ContainerName
	tty := option.Tty
	volumeMapping := option.VolumeMapping
	portMappings := option.PortMappings
	envs := option.Envs

	containerId := util.GenRandStrBytes(10)
	if len(containerName) == 0 {
		containerName = containerId
	}

	parent, wPipe, err := container.NewParentProcess(tty, volumeMapping, imageName, containerName, envs)
	if err != nil {
		return fmt.Errorf("NewParentProcess() error %v", err)
	}
	if err := parent.Start(); err != nil {
		return fmt.Errorf("Start() parent process error %v", err)
	}

	// Parent process in the container should wait here to read piped command.

	// TODO: recordContainerInfo
	if err := writeContainerInfo(parent.Process.Pid, cmdArr, portMappings, containerName, containerId, volumeMapping); err != nil {
		return fmt.Errorf("recordContainerInfo() error %v", err)
	}
	// TODO: NewCgroupManager
	// TODO: config container network

	// Send init command to parent process.
	if err := sendInitCommand(cmdArr, wPipe); err != nil {
		log.Errorf("sendInitCommand() %v error %v", cmdArr, err)
	}

	if err := wPipe.Close(); err != nil {
		log.Error("Close() error %v", err)
	}

	if tty {
		// Here the current tty of this process is piped to the parent process.
		// Need to wait for the termination of the parent process.
		// ExitError: The command fails to execute or doesn't complete successfully.
		if err := parent.Wait(); err != nil {
			log.Errorf("Wait() error %v", err)
		}
		// deleteContainerInfo
		if err := container.DeleteWorkspace(volumeMapping, containerName); err != nil {
			log.Errorf("DeleteWorkspace() volume mapping %s and container name %s error %v. You may need to delete something manually", volumeMapping, containerName, err)
		}
	}

	return nil
}

func sendInitCommand(cmdArr []string, wPipe *os.File) error {
	command := strings.Join(cmdArr, " ")
	log.Infof("Full init command is %s", command)
	if _, err := wPipe.WriteString(command); err != nil {
		return fmt.Errorf("WriteString() %s error %v", command, err)
	}

	return nil
}

func writeContainerInfo(containerPid int, cmdArr, portMappings []string, name, id, volumeMapping string) error {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(cmdArr, " ")
	containerInfo := &container.ContainerInfo{
		Pid: containerPid,
		Id: id,
		Name: name,
		Command: command,
		CreateTime: createTime,
		Status: container.STATUS_RUNNING,
		VolumeMapping: volumeMapping,
		PortMappings: portMappings,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("Marshal() %v error %v", containerInfo, err)
	}

	jsonStr := string(jsonBytes)

	dirPath := filepath.Join(container.DEFAULT_INFO_DIR_PATH, name)
	if err := os.MkdirAll(dirPath, 0622); err != nil {
		return fmt.Errorf("MkdirAll() %s error %v", dirPath, err)
	}
	filePath := filepath.Join(dirPath, container.CONFIG_FILE_NAME)
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Create() %s error %v", filePath, err)
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		return fmt.Errorf("WriteString() error %v", jsonStr)
	}

	return nil
}
