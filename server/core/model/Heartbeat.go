package model

import (
	"os"

	"github.com/Earthmark/Motley/server/core/status"
	"github.com/Earthmark/Motley/server/gen"
)

func (m *Manager) Update() {
	manager := status.Proc(os.Getpid())
	servers := make([]gen.Server, 0)
	for id, s := range m.servers {
		if s.p != nil {
			s.p.Enforce()
		}
		servers = append(servers, gen.Server{
			ID:      id,
			Name:    id,
			Options: s.conf,
			Status:  s.status(),
		})
	}
	status := gen.Status{
		Manager: gen.ManagerStatus{
			Pid:        manager.Pid,
			UsedMemory: manager.UsedMemory,
			CPULoad:    manager.CPULoad,
		},
		System:  status.System(),
		Servers: servers,
	}

	m.Status = status
}
