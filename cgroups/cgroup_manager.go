package cgroups

import (
	log "github.com/sirupsen/logrus"
)

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
	for _, subsystem := range subsystems {
		if err := subsystem.Apply(c.Path, pid); err != nil {
			log.Errorf("Apply() cgroup %s to pid %d error %v", c.Path, pid, err)
		}
	}

	return nil
}

// Set resource limitations of this cgroup.
func (c *CgroupManager) Set(res *ResourceConfig) error {
	for _, subsystem := range subsystems {
		if err := subsystem.Set(c.Path, res); err != nil {
			log.Errorf("Set() cgroup %s error %v", c.Path, err)
		}
	}

	return nil
}

// Release this cgroup.
func (c *CgroupManager) Destroy() error {
	for _, subsystem := range subsystems {
		if err := subsystem.Remove(c.Path); err != nil {
			log.Errorf("Remove() cgroup %s error %v", c.Path, err)
		}
	}

	return nil
}
