package cgroups

import "fmt"
import log "github.com/sirupsen/logrus"

type CgroupManager struct {
	Path string // The path of this cgroup in the overall hierarchy.
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Add process of pid to this cgroup.
func (c *CgroupManager) Apply(pid int) error {
	var retErr error
	for _, subsystem := range subsystems {
		if err := subsystem.Apply(c.Path, pid); err != nil {
			retErr = fmt.Errorf("Apply() cgroup %s to pid %d error %v", c.Path, pid, err)
			log.Error(retErr)
		}
	}

	return retErr
}

// Set resource limitations of this cgroup.
func (c *CgroupManager) Set(res *ResourceConfig) error {
	var retErr error
	for _, subsystem := range subsystems {
		if err := subsystem.Set(c.Path, res); err != nil {
			retErr = fmt.Errorf("Set() cgroup %s error %v", c.Path, err)
			log.Error(retErr)
		}
	}

	return retErr
}

// Release this cgroup.
func (c *CgroupManager) Destroy() error {
	var retErr error
	for _, subsystem := range subsystems {
		if err := subsystem.Remove(c.Path); err != nil {
			retErr = fmt.Errorf("Remove() cgroup %s error %v", c.Path, err)
			log.Error(retErr)
		}
	}

	return retErr
}

