package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/chengzeyi/dicker/command"
)

const (
	RUNNING                 = "runing"
	STOPPED                 = "stopped"
	EXITED                  = "exited"
	DEFAULT_INFO_DIR_PATH   = "/var/run/dicker"
	CONDIG_FILE_NAME        = "config.json"
	CONTAINER_LOG_FILE_NAME = "container.log"
	ROOT_DIR_PATH           = "/root"
	MNT_DIR_PATH            = "/root/mnt"
	WRITE_LAYER_DIR_PATH    = "/root/wirte_layer"
)

type ContainerInfo struct {
	Pid         string   `json:"pid"`          // Container init process's pid on the host OS.
	Id          string   `json:"id"`           // Container id.
	Name        string   `json:"name"`         // Container name.
	Command     string   `json:"command"`      // Container init command.
	CreateTime  string   `json:"create_time"`  // Container created time.
	Status      string   `json:"status"`       // Container status description.
	Volume      string   `json:"volume"`       // Container data volume.
	PortMappint []string `json:"port_mapping"` // Container port mapping.
}

func NewParentProcess(tty bool, containerName, volume, imageName string, envs []string) (*exec.Cmd, *os.File, error) {
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Pipe error %v", err)
	}
	selfCmd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return nil, nil, fmt.Errorf("Readlink /proc/self/exe error %v", err)
	}

	initCmd := exec.Command(selfCmd, command.COMMAND_INIT)
	// Cloneflags contains all the namespace flags except CLONE_NEWUSER
	// CLONE_NEWUTS: In the new UTS namespace.
	// CLONE_NEWPID: In the new pid namespace.
	// CLONE_NEWNS: In the new mount namespace.
	// CLONE_NEWNET: In the new net namespace.
	// CLONE_NEWIPC: In the new ipc namespace.
	initCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		initCmd.Stdin = os.Stdin
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
	} else {
		dirPath := filepath.Join(DEFAULT_INFO_DIR_PATH, containerName)
		if err := os.MkdirAll(dirPath, 0622); err != nil {
			return nil, nil, fmt.Errorf("MkdirAll %s error %v", dirPath, err)
		}
		stdOutLogFilePath := filepath.Join(dirPath, CONTAINER_LOG_FILE_NAME)
		stdOutLogFile, err := os.Create(stdOutLogFilePath)
		if err != nil {
			return nil, nil, fmt.Errorf("Create %s error %v", stdOutLogFilePath, err)
		}
		initCmd.Stdout = stdOutLogFile
	}

	initCmd.ExtraFiles = []*os.File{
		rPipe,
	}
	initCmd.Env = append(os.Environ(), envs...)
	initCmd.Dir = filepath.Join(MNT_DIR_PATH, containerName)

	// TODO: NewWorkspace

	return initCmd, wPipe, nil
}
