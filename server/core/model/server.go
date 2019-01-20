package model

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/Earthmark/Motley/server/config"
	"github.com/Earthmark/Motley/server/core/proc"
	"github.com/Earthmark/Motley/server/core/status"
	"github.com/Earthmark/Motley/server/gen"
)

// TODO: Add RCON channel and other stuff.
type server struct {
	exec    string
	conf    *config.ServerOptions
	pLock   sync.Mutex
	p       *proc.EnforcedProcess
	updated func()
}

func bindServer(exec string, conf *config.ServerOptions, updatedCallback func()) *server {
	s := &server{
		exec:    exec,
		conf:    conf,
		pLock:   sync.Mutex{},
		updated: updatedCallback,
	}
	if conf.Pid != 0 {
		s.launchManager(conf.Pid)
	}
	return s
}

func (s *server) status() *gen.ServerStatus {
	if s.p != nil {
		pid, err := s.p.Enforce()
		if err != nil {
			log.Printf("Error launching server, %v", err)
		}
		if s.conf.Pid != s.p.Pid {
			s.conf.Pid = s.p.Pid
			s.updated()
		}
		b := status.Proc(pid)
		return &gen.ServerStatus{
			Players:    -1,
			MaxPlayers: -1,
			Pid:        int(pid),
			CPULoad:    b.CPULoad,
			UsedMemory: b.UsedMemory,
		}
	}
	return nil
}

func itoa(num int) string {
	return strconv.Itoa(num)
}

func btos(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (s *server) launchManager(prevPid int) {
	arg1Sections := map[string]string{
		"ServerX":              itoa(s.conf.ServerX),
		"ServerY":              itoa(s.conf.ServerY),
		"Port":                 itoa(s.conf.Port),
		"AltSaveDirectoryName": s.conf.AltSaveDirectoryName,
		"MaxPlayers":           itoa(s.conf.MaxPlayers),
		"ReservedPlayerSlots":  itoa(s.conf.ReservedPlayerSlots),
		"SeamlessIP":           s.conf.SeamlessIP,
	}
	arg1joined := make([]string, 0)
	for k, v := range arg1Sections {
		arg1joined = append(arg1joined, fmt.Sprintf("%s=%s", k, v))
	}

	args := []string{
		fmt.Sprintf("ocean?%s%s", strings.Join(arg1joined, "?"), s.conf.PreExecArgs),
		"-game",
		"-server",
		"-log",
		"-NoCrashDialog",
	}
	if !s.conf.BattleEye {
		args = append(args, "-NoBattlEye")
	}
	if len(s.conf.PostExecArgs) > 0 {
		args = append(args, s.conf.PostExecArgs...)
	}

	s.p = proc.ManageProcess(s.exec, args, prevPid)
}
