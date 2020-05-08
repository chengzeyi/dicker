package command

import (
	"os"
	"os/exec"
	"syscall"
)

type RunOption struct {
	Tty bool
}

func Run(option *RunOption, imageName string, cmdArr []string) error {
	return nil
}

func NewParentProcess(tty bool, cmd string) *exec.Cmd {
	args := []string{"init", cmd}
	command := exec.Command("/proc/self/exe", args...)
	command.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	return command
}
