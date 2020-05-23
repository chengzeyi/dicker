package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const PIVOT_PUT_OLD_DIR_NAME = ".pivot_put_old"

func RunContainerInitProcess() error {
	// Get the command to be executed.
	cmdArr := readUserCommand()
	if cmdArr == nil || len(cmdArr) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArr is nil")
	}

	mount()

	path, err := exec.LookPath(cmdArr[0])
	if err != nil {
		log.Errorf("LookPath error %v", err)
		return err
	}
	log.Printf("Find path %s\n", path)
	if err := syscall.Exec(path, cmdArr[0:], os.Environ()); err != nil {
		log.Errorf("%s", err.Error())
	}

	return nil
}

func readUserCommand() []string {
	// 3 is the file descriptor after 0(stdin), 1(stdout) and 2(stderr).
	pipe := os.NewFile(3, "pipe")
	defer pipe.Close()

	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("Init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func mount() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Get working directory error %v", err)
	}
	log.Infof("Working directory is %s", wd)

	// After this, the working directory becomes '/'.
	if err := pivotRoot(wd); err != nil {
		return fmt.Errorf("pivotRoot() %s error %v", wd, err)
	}

	//_MS_NOEXEC: Do not allow program to be executed from this filesystem.
	// MS_NO_SUID: Do not honor set-user-ID and set-group-ID bits or file capabilities when executing programs from this filesystem.
	// MS_NODEV: Do not allow access to devices (special files) on this filesystem.
	if err := syscall.Mount("proc", "/proc", "proc", syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV, ""); err != nil {
		return fmt.Errorf("Mount() proc to /proc error %v", err)
	}
	// MS_STRICTATIME: Always update the last access time (atime) when files on this filesystem are accessed.
	// MS_RELATIME: Only update atime if it is less than or equal to mtime or ctime. (default behaviour since Linux 2.6.20)
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf("Mount() tmpfs to /dev error %v", err)
	}

	return nil
}

// Move the root filesystem to the directory 'root'
func pivotRoot(root string) error {
	// The new root and old root should not in the same file system.
	// By binding, the remaining bits other than MS_REC in flags are ignored.
	// fstype and data are also ignored.
	// This creates a security boundary for certain operations like hard link.
	if err := syscall.Mount(root, root, "", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount() %s to itself error %v", root, err)
	}
	// .pivot_root is used for storing the old root.
	putold := filepath.Join(root, PIVOT_PUT_OLD_DIR_NAME)
	// It should not exist before.
	if err := os.Mkdir(putold, 0777); err != nil {
		return fmt.Errorf("Mkdir() %s error %v", PIVOT_PUT_OLD_DIR_NAME, err)
	}
	// Make root to be / and put the old root in .pivot_root.
	if err := syscall.PivotRoot(root, putold); err != nil {
		return fmt.Errorf("PivotRoot() error %v", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("Chdir() / error %v", err)
	}
	putold = filepath.Join("/", PIVOT_PUT_OLD_DIR_NAME)
	// Unmount the old '/'.
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("Unmount() directory %s error %v", PIVOT_PUT_OLD_DIR_NAME, err)
	}
	// Remove the tmp file.
	if err := os.Remove(putold); err != nil {
		return fmt.Errorf("Remove() directory %s error %v", PIVOT_PUT_OLD_DIR_NAME, err)
	}

	return nil
}
