package paralleldl

import (
	"context"
	"sync"
	"sync/atomic"
)

type results struct {
	wg         sync.WaitGroup
	errMax     int64
	errCounts  int64
	errChannel chan error
}

func newResults(maxQueues int, maxErrRequests int64) *errRequests {
	return &errRequests{
		max:     maxErrRequests,
		counts:  0,
		channel: make(chan error, maxQueues),
	}
}

func (e *errRequests) add() {
	atomic.AddInt64(&e.counts, 1)
}

func (e *errRequests) isOver() bool {
	if e.max == 0 {
		return false
	}
	return atomic.LoadInt64(&e.counts) < e.max
}

type dispatcher struct {
	jobQueue    chan string
	workers     []*worker
	workerPool  chan *worker
	done        chan struct{}
	errRequests *errRequests
}

func newDispatcher(c *Client, maxQueues, maxWorkers int) *dispatcher {
	errReq := newErrRequests(maxQueues, c.opt.MaxErrorRequests)
	d := &dispatcher{
		jobQueue:    make(chan string, maxQueues),
		workerPool:  make(chan *worker, maxWorkers),
		errRequests: errReq,
	}
	d.workers = make([]*worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		w := worker{
			pool:        d.workerPool,
			client:      c,
			target:      make(chan string),
			done:        make(chan struct{}),
			errRequests: errReq,
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

	go func() {
		for {
			select {
			case v := <-d.jobQueue:
				worker := <-d.workerPool
				worker.target <- v
			case <-d.errRequests.channel:
				d.errRequests.add()
				if d.errRequests.isOver() {
					ctxCancel()
					return
				}
			case <-d.done:
				return
			}
		}
	}()
}

func (d *dispatcher) stop() {
	wg.Wait()
	for _, w := range d.workers {
		w.stop()
	}
}

func (d *dispatcher) add(v string) {
	wg.Add(1)
	d.jobQueue <- v
}

type worker struct {
	pool        chan *worker
	errRequests *errRequests
	client      *Client
	done        chan struct{}
	target      chan string
}

func (w *worker) start(ctx context.Context) {
	go func() {
		for {
			w.pool <- w
			select {
			case target := <-w.target:
				err := w.client.download(ctx, target)
				if err != nil {
					w.errRequests.channel <- err
				}
				wg.Done()

			case <-w.done:
				return
			}
		}
	}()
}

func (w *worker) stop() {
	close(w.done)
}
