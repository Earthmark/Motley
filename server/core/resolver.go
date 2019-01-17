package core

//go:generate go run ../scripts/gqlgen.go -v
//go:generate vfsgendev -source="github.com/Earthmark/Motley/server/core".Client

import (
	"context"
	"sync"
	"time"

	"github.com/Earthmark/Motley/server/config"
	"github.com/Earthmark/Motley/server/gen"
)

type resolver struct {
	configPath string
	config     config.Config

	statusTicker        *time.Ticker
	statusListenersLock sync.Mutex
	statusListenerIdx   int64
	statusListeners     map[int64]chan gen.Status
	status              gen.Status
}

func CreateResolver(configSource string, conf config.Config) (gen.ResolverRoot, error) {
	resol := &resolver{
		configPath:          configSource,
		config:              conf,
		statusTicker:        time.NewTicker(time.Second * 1),
		statusListenersLock: sync.Mutex{},
		statusListenerIdx:   0,
		statusListeners:     make(map[int64]chan gen.Status),
	}

	for _, options := range conf.Servers {
		resol.addServer(options)
	}

	go func() {
		for range resol.statusTicker.C {
			status := UpdateSystemMetrics()
			resol.status = status
			resol.statusListenersLock.Lock()
			for _, ch := range resol.statusListeners {
				ch <- status
			}
			resol.statusListenersLock.Unlock()
		}
	}()

	return resol, nil
}

func (r *resolver) Query() gen.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct {
	r *resolver
}

func (r *resolver) Subscription() gen.SubscriptionResolver {
	return &subscriptionResolver{r}
}

type subscriptionResolver struct {
	r *resolver
}

func (r *resolver) addServer(options config.ServerOptions) {

}

func (q *subscriptionResolver) Status(ctx context.Context) (<-chan gen.Status, error) {
	statusChan := make(chan gen.Status, 1)

	q.r.statusListenersLock.Lock()
	idx := q.r.statusListenerIdx
	q.r.statusListenerIdx = idx + 1
	q.r.statusListeners[idx] = statusChan
	q.r.statusListenersLock.Unlock()
	go func() {
		<-ctx.Done()
		q.r.statusListenersLock.Lock()
		delete(q.r.statusListeners, idx)
		q.r.statusListenersLock.Unlock()
	}()

	return statusChan, nil
}

func (q *queryResolver) System(ctx context.Context) (gen.SystemStatus, error) {
	return q.r.status.System, nil
}

func (q *queryResolver) Manager(ctx context.Context) (gen.ManagerStatus, error) {
	return q.r.status.Manager, nil
}

func (q *queryResolver) Server(ctx context.Context, id string) (*gen.Server, error) {
	panic("not implemented")
}
func (q *queryResolver) Servers(ctx context.Context) ([]gen.Server, error) {
	panic("not implemented")
}
