package cgroups

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

type SubsystemBase struct {}

type CpuSubsystem struct {}

func (CpuSubsystem) Name() string {

}

func (CpuSubsystem) Set(path string, res *ResourceConfig) error {
	panic("not implemented") // TODO: Implement
}

func (CpuSubsystem) Apply(path string, pid int) error {
	panic("not implemented") // TODO: Implement
}

func (CpuSubsystem) Remove(path string) error {
	panic("not implemented") // TODO: Implement
}

