package paralleldl

import (
	"context"
	"sync"
	"sync/atomic"
)

var wg sync.WaitGroup

type dispatcher struct {
	jobQueue     chan string
	workers      []*worker
	done         chan struct{}
	maxErrCounts int64
	errCounts    int64
}

func newDispatcher(c *Client, maxQueues, maxWorkers int) *dispatcher {
	d := &dispatcher{
		jobQueue:     make(chan string, maxQueues),
		maxErrCounts: c.opt.MaxErrorRequests,
		done:         make(chan struct{}, 1),
	}
	d.workers = make([]*worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		w := worker{
			client: c,
			done:   make(chan struct{}, 1),
		}
		d.workers[i] = &w
	}
	return d
}

func (d *dispatcher) start(list []string) {
	for _, v := range list {
		d.jobQueue <- v
	}

	errChannel := make(chan error)

	ctx, ctxCancel := context.WithCancel(context.Background())
	wg.Add(len(d.workers))
	for _, w := range d.workers {
		w.start(ctx, d.jobQueue, errChannel)
	}

	go func() {
		for {
			select {
			case err := <-errChannel:
				if err != nil {
					atomic.AddInt64(&d.errCounts, 1)
					if d.maxErrCounts != 0 && d.maxErrCounts <= atomic.LoadInt64(&d.errCounts) {
						ctxCancel()
						d.stop()
						return
					}
				}
			case <-d.done:
				return
			}
		}
	}()
}

func (d *dispatcher) wait() {
	close(d.jobQueue)
	wg.Wait()
	d.done <- struct{}{}
}

func (d *dispatcher) stop() {
	for _, w := range d.workers {
		w.stop()
	}
}

type worker struct {
	client *Client
	done   chan struct{}
}

func (w *worker) start(ctx context.Context, jobQueue chan string, errChannel chan error) {
	go func() {
		defer wg.Done()
		for target := range jobQueue {
			err := w.client.download(ctx, target)
			select {
			case errChannel <- err:
			case <-w.done:
				return
			}
		}
	}()
}

func (w *worker) stop() {
	close(w.done)
}
