package status

import (
	"code.cloudfoundry.org/bytefmt"
	"github.com/Earthmark/Motley/server/gen"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

func System() gen.SystemStatus {
	system := gen.SystemStatus{
		UsedMemory:  -1,
		TotalMemory: -1,
		CPULoad:     -1.0,
	}
	if mem, err := mem.VirtualMemory(); err == nil {
		system.UsedMemory = int(mem.Used / bytefmt.MEGABYTE)
		system.TotalMemory = int(mem.Total / bytefmt.MEGABYTE)
	}
	if proc, err := cpu.Percent(0, false); err == nil {
		system.CPULoad = proc[0]
	}
	return system
}

func Proc(pid int) *ProcStatus {
	if proc, err := process.NewProcess(int32(pid)); err == nil {
		status := &ProcStatus{
			Pid:        int(pid),
			UsedMemory: -1,
			CPULoad:    -1.0,
		}
		if mem, err := proc.MemoryInfo(); err == nil {
			status.UsedMemory = int(mem.RSS / bytefmt.MEGABYTE)
		} else {
			status.UsedMemory = -1
		}

		if perc, err := proc.Percent(0); err == nil {
			status.CPULoad = perc
		} else {
			status.CPULoad = -1.0
		}
		return status
	}
	return nil
}

type ProcStatus struct {
	Pid        int
	CPULoad    float64
	UsedMemory int
}
