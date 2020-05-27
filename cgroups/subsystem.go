package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/chengzeyi/dicker/util"
)

type ResourceConfig struct {
	MemoryLimit string
	CpuShares   string
	Cpuset      string
}

var subsystems = []Subsystem{
	&CpuSubsystem{},
	&CpusetSubsystem{},
	&MemorySubsystem{},
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

type SubsystemBase struct{}

type CpuSubsystem struct {
	SubsystemBase
}

type CpusetSubsystem struct {
	SubsystemBase
}

type MemorySubsystem struct {
	SubsystemBase
}

func (s *SubsystemBase) Name() string {
	panic("not implemented")
}

func (s *SubsystemBase) set(path, key, val string) error {
	if len(val) == 0 {
		return nil
	}

	cgroupPath, err := util.GetCgroupPath(s.Name(), path, true)
	if err != nil {
		return fmt.Errorf("GetCgroupPath() of subsystem %s error %v", s.Name(), err)
	}

	filePath := filepath.Join(cgroupPath, key)
	if err := ioutil.WriteFile(filePath, []byte(val), 0644); err != nil {
		return fmt.Errorf("WriteFile() %s error %v", filePath, err)
	}

	return nil
}

func (s *SubsystemBase) Apply(path string, pid int) error {
	cgroupPath, err := util.GetCgroupPath(s.Name(), path, false)
	if err != nil {
		return fmt.Errorf("GetCgroupPath() of subsystem %s error %v", s.Name(), err)
	}

	tasksFilePath := filepath.Join(cgroupPath, "tasks")
	if err := ioutil.WriteFile(tasksFilePath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("WriteFile %s error %v", tasksFilePath, err)
	}

	return nil
}

func (s *SubsystemBase) Remove(path string) error {
	cgroupPath, err := util.GetCgroupPath(s.Name(), path, false)
	if err != nil {
		return fmt.Errorf("GetCgroupPath() of subsystem %s error %v", s.Name(), err)
	}
	if err := os.RemoveAll(cgroupPath); err != nil {
		return fmt.Errorf("RemoveAll() %s error %v", cgroupPath, err)
	}

	return nil
}

func (s *CpuSubsystem) Name() string {
	return "cpu"
}

func (s *CpuSubsystem) Set(path string, res *ResourceConfig) error {
	return s.set(path, "cpu.shares", res.CpuShares)
}

func (s *CpusetSubsystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubsystem) Set(path string, res *ResourceConfig) error {
	return s.set(path, "cpuset.cpus", res.Cpuset)
}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

func (s *MemorySubsystem) Set(path string, res *ResourceConfig) error {
	return s.set(path, "memory.limit_in_bytes", res.MemoryLimit)
}
