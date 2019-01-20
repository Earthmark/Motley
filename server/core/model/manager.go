package model

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/Earthmark/Motley/server/config"
	"github.com/Earthmark/Motley/server/core/status"
	"github.com/Earthmark/Motley/server/gen"
)

type Manager struct {
	conf *config.Config

	Status gen.Status

	serverLock sync.Mutex
	servers    map[string]*server
}

func Create(conf *config.Config) *Manager {
	m := &Manager{
		conf:       conf,
		serverLock: sync.Mutex{},
		servers:    make(map[string]*server),
	}

	for i, c := range m.conf.Servers {
		m.Add(i, c)
	}

	return m
}

func (m *Manager) managerStatus() gen.ManagerStatus {
	manager := status.Proc(os.Getpid())
	return gen.ManagerStatus{
		Pid:        manager.Pid,
		UsedMemory: manager.UsedMemory,
		CPULoad:    manager.CPULoad,
	}
}

func (m *Manager) serverStatus() []gen.Server {
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
	return servers
}

func (m *Manager) Update() gen.Status {
	status := gen.Status{
		Manager: m.managerStatus(),
		System:  status.System(),
		Servers: m.serverStatus(),
	}
	m.Status = status
	return status
}

func (m *Manager) Start(id string) error {
	m.serverLock.Lock()
	defer m.serverLock.Unlock()
	if s, ok := m.servers[id]; ok {
		if s.p == nil {
			s.launchManager(0)
		}
		return nil
	}
	return errors.New("server with the provided ID was not found")
}

func (m *Manager) Stop(id string) error {
	m.serverLock.Lock()
	defer m.serverLock.Unlock()
	if s, ok := m.servers[id]; ok {
		if s.p == nil {
			return nil
		}
		if err := s.p.Kill(); err != nil {
			return err
		}
		s.p = nil
		return nil
	}
	return errors.New("server with the provided ID was not found")
}

func (m *Manager) Add(id string, conf config.ServerOptions) error {
	m.serverLock.Lock()
	defer m.serverLock.Unlock()
	if _, ok := m.servers[id]; ok {
		return errors.New("server with this ID already exists")
	}
	m.servers[id] = bindServer(m.conf.ShooterGame, conf, func() {
		if err := m.conf.Save(); err != nil {
			log.Printf("Failed to save after server %s was added, the server may not persist if Motley is restarted: %v", id, err)
		}
	})
	return nil
}
