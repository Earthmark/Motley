package core

import (
	"os"

	"code.cloudfoundry.org/bytefmt"
	"github.com/Earthmark/Motley/server/gen"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

func defaultManager(pid int) gen.ManagerStatus {
	return gen.ManagerStatus{
		Pid:        pid,
		UsedMemory: -1,
		CPULoad:    -1.0,
	}
}

func defaultSystem() gen.SystemStatus {
	return gen.SystemStatus{
		UsedMemory:  -1,
		TotalMemory: -1,
	}
}

func UpdateSystemMetrics() gen.Status {
	pid := os.Getpid()

	manager := defaultManager(pid)
	if proc, err := process.NewProcess(int32(pid)); err == nil {
		if mem, err := proc.MemoryInfo(); err == nil {
			manager.UsedMemory = int(mem.RSS / bytefmt.MEGABYTE)
		} else {
			manager.UsedMemory = -1
		}

		if perc, err := proc.Percent(0); err == nil {
			manager.CPULoad = perc
		} else {
			manager.CPULoad = -1.0
		}

	}

	system := defaultSystem()
	if mem, err := mem.VirtualMemory(); err == nil {
		system.UsedMemory = int(mem.Used / bytefmt.MEGABYTE)
		system.TotalMemory = int(mem.Total / bytefmt.MEGABYTE)
	}

	return gen.Status{
		Manager: manager,
		System:  system,
	}
}
