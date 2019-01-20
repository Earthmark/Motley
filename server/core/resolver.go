package core

//go:generate go run ../scripts/gqlgen.go -v
//go:generate vfsgendev -source="github.com/Earthmark/Motley/server/core".Client

import (
	"context"
	"sync"
	"time"

	"github.com/Earthmark/Motley/server/config"
	"github.com/Earthmark/Motley/server/core/model"
	"github.com/Earthmark/Motley/server/gen"
)

type resolver struct {
	m *model.Manager
	t *time.Ticker

	s *subscriptionResolver
}

type subscriptionResolver struct {
	r                   *resolver
	statusListenersLock sync.Mutex
	statusListenerIdx   int64
	statusListeners     map[int64]chan gen.Status
}

func CreateResolver(conf *config.Config) gen.ResolverRoot {
	r := &resolver{
		m: model.Create(conf),
		t: time.NewTicker(time.Duration(conf.StatusRateSeconds) * time.Second),
		s: &subscriptionResolver{
			statusListenersLock: sync.Mutex{},
			statusListenerIdx:   0,
			statusListeners:     make(map[int64]chan gen.Status),
		},
	}

	go r.updateLoop()

	return r
}

func (r *resolver) updateLoop() {
	for range r.t.C {
		r.update()
	}
}

func (r *resolver) update() {
	status := r.m.Update()
	r.s.statusListenersLock.Lock()
	defer r.s.statusListenersLock.Unlock()
	for _, c := range r.s.statusListeners {
		c <- status
	}
}

func (r *resolver) Query() gen.QueryResolver {
	return r
}

type queryResolver struct {
	r *resolver
}

func (r *resolver) Subscription() gen.SubscriptionResolver {
	return r.s
}

func (r *resolver) addServer(options config.ServerOptions) {

}

func (q *subscriptionResolver) Status(ctx context.Context) (<-chan gen.Status, error) {
	statusChan := make(chan gen.Status, 1)

	q.statusListenersLock.Lock()
	idx := q.statusListenerIdx
	q.statusListenerIdx = idx + 1
	q.statusListeners[idx] = statusChan
	q.statusListenersLock.Unlock()
	go func() {
		<-ctx.Done()
		q.statusListenersLock.Lock()
		delete(q.statusListeners, idx)
		q.statusListenersLock.Unlock()
	}()

	return statusChan, nil
}

func (r *resolver) Status(ctx context.Context) (gen.Status, error) {
	return r.m.Status, nil
}
