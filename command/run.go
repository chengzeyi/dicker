package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/chengzeyi/dicker/container"
	"github.com/chengzeyi/dicker/util"

	log "github.com/sirupsen/logrus"
)

type RunOption struct {
	Tty bool
	ContainerName string
	VolumeMapping string
	Envs []string
}

func Run(option *RunOption, imageName string, cmdArr []string) error {
	containerName := option.ContainerName
	tty := option.Tty
	volumeMapping := option.VolumeMapping
	envs := option.Envs

	containerId := util.GenRandStrBytes(10)
	if len(containerName) == 0 {
		containerName = containerId
	}

	parent, wPipe, err := container.NewParentProcess(tty, volumeMapping, imageName, containerName, envs)
	if err != nil {
		return fmt.Errorf("NewParentProcess error %v", err)
	}
	if err := parent.Start(); err != nil {
		return fmt.Errorf("Parent process Start error %v", err)
	}

	// Parent process should wait here to read piped command.


	// TODO: recordContainerInfo
	// TODO: NewCgroupManager
	// TODO: config container network

	if err := sendInitCommand(cmdArr, wPipe); err != nil {
		log.Errorf("sendInitCommand %v error %v", cmdArr, err)
	}

	if err := wPipe.Close(); err != nil {
		log.Error("Close error %v", err)
	}

	if tty {
		// Here the current tty of this process is piped to the parent process.
		// Need to wait for the termination of the parent process.
		// ExitError: The command fails to execute or doesn't complete successfully.
		if err := parent.Wait(); err != nil {
			log.Errorf("Wait error %v", err)
		}
		// deleteContainerInfo
		if err := container.DeleteWorkspace(volumeMapping, containerName); err != nil {
			log.Errorf("DeleteWorkspace volume mapping %s and container name %s error %v. You may need to delete something manually", volumeMapping, containerName, err)
		}
	}

	return nil
}

func sendInitCommand(cmdArr []string, wPipe *os.File) error {
	command := strings.Join(cmdArr, " ")
	log.Infof("Full init command is %s", command)
	if _, err := wPipe.WriteString(command); err != nil {
		return fmt.Errorf("WriteString %s error %v", command, err)
	}

	return nil
}
