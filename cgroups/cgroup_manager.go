package cgroups


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
	return nil
}

// Set resource limitations of this cgroup.
func (c *CgroupManager) Set(pid int) error {
	return nil
}

// Release this cgroup.
func (c *CgroupManager) Destroy(pid int) error {
	return nil
}

