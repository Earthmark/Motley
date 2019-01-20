package proc

import (
	"os/exec"

	"github.com/shirou/gopsutil/process"
)

// EnforcedProcess - A wrapper that will keep enforcing that a process is in a requested state.
// It will either start or stop the process to achieve that state.
type EnforcedProcess struct {
	Pid       int
	command   string
	arguments []string
}

func ManageProcess(command string, arguments []string, currentPid int) *EnforcedProcess {
	return &EnforcedProcess{
		Pid: currentPid,

		command:   command,
		arguments: arguments,
	}
}

func (p *EnforcedProcess) Kill() error {
	if proc, err := process.NewProcess(int32(p.Pid)); err == nil {
		if err := proc.Kill(); err != nil {
			return err
		}
	}
	return nil
}

func (p *EnforcedProcess) Enforce() (int, error) {
	var err error
	isRunning := false
	if p.Pid != 0 {
		isRunning, err = process.PidExists(int32(p.Pid))
		if err != nil {
			return 0, err
		}
		// If it's not actually running, remove our record that it's running.
		if !isRunning {
			p.Pid = 0
		}
	}

	if !isRunning {
		command := exec.Command(p.command, p.arguments...)
		if err := command.Start(); err != nil {
			return 0, err
		}
		p.Pid = command.Process.Pid
	}
	return p.Pid, nil
}
