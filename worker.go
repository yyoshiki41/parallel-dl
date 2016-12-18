package paralleldl

import (
	"context"
	"sync"
)

var (
	wg sync.WaitGroup
)

type dispatcher struct {
	jobQueue   chan string
	workers    []*worker
	workerPool chan *worker
}

func newDispatcher(c *Client, maxQueues, maxWorkers int) *dispatcher {
	d := &dispatcher{
		jobQueue:   make(chan string, maxQueues),
		workerPool: make(chan *worker, maxWorkers),
	}
	d.workers = make([]*worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		w := worker{
			pool:   d.workerPool,
			client: c,
			target: make(chan string),
			quit:   make(chan struct{}),
		}
		d.workers[i] = &w
	}
	return d
}

func (d *dispatcher) start() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	for _, w := range d.workers {
		w.start(ctx)
	}

	maxErrRequests := int(d.workers[0].client.opt.MaxErrorRequests)
	go func() {
		defer ctxCancel()
		for {
			select {
			case v := <-d.jobQueue:
				worker := <-d.workerPool
				worker.target <- v
			case <-errChannel:
				// Not accurate, because errChannel is a global variable.
				if maxErrRequests != 0 && len(errChannel) >= maxErrRequests {
					ctxCancel()
					d.stop()
				}
			}
		}
	}()
}

func (d *dispatcher) stop() {
	for _, w := range d.workers {
		w.stop()
	}
}

func (d *dispatcher) wait() {
	wg.Wait()
}

func (d *dispatcher) add(v string) {
	wg.Add(1)
	d.jobQueue <- v
}

type worker struct {
	pool   chan *worker
	client *Client
	target chan string
	quit   chan struct{}
}

func (w *worker) start(ctx context.Context) {
	go func() {
		for {
			w.pool <- w
			select {
			case target := <-w.target:
				err := w.client.download(ctx, target)
				if err != nil {
					errChannel <- err
				}
				wg.Done()

			case <-w.quit:
				return
			}
		}
	}()
}

func (w *worker) stop() {
	close(w.quit)
}
