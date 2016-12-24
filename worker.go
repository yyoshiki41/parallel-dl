package paralleldl

import (
	"context"
	"sync"
	"sync/atomic"
)

type results struct {
	wg     sync.WaitGroup
	errCh  chan error
	errCnt int64
}

type dispatcher struct {
	jobQueue chan string
	workers  []*worker
	res      *results
	done     chan struct{}
}

func newDispatcher(c *Client, maxQueues, maxWorkers int) *dispatcher {
	r := &results{errCh: make(chan error)}

	d := &dispatcher{
		jobQueue: make(chan string, maxQueues),
		workers:  make([]*worker, maxWorkers),
		res:      r,
		done:     make(chan struct{}, 1),
	}
	for i := 0; i < maxWorkers; i++ {
		w := worker{
			client: c,
			done:   make(chan struct{}, 1),
			res:    r,
		}
		d.workers[i] = &w
	}
	return d
}

func (d *dispatcher) start(list []string) {
	maxErrCounts := d.workers[0].client.opt.MaxErrorRequests
	d.res.wg.Add(len(d.workers))

	for _, v := range list {
		d.jobQueue <- v
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	for _, w := range d.workers {
		w.start(ctx, d.jobQueue)
	}

	go func() {
		for {
			select {
			case err := <-d.res.errCh:
				if err != nil {
					if maxErrCounts != 0 {
						continue
					}
					if maxErrCounts <= atomic.LoadInt64(&d.res.errCnt) {
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
	d.res.wg.Wait()
	d.done <- struct{}{}
}

func (d *dispatcher) stop() {
	for _, w := range d.workers {
		w.stop()
	}
}

func (d *dispatcher) errCounts() int64 {
	return atomic.LoadInt64(&d.res.errCnt)
}

type worker struct {
	client *Client
	res    *results
	done   chan struct{}
}

func (w *worker) start(ctx context.Context, jobQueue chan string) {
	go func() {
		defer w.res.wg.Done()

		for target := range jobQueue {
			err := w.client.download(ctx, target)
			if err != nil {
				atomic.AddInt64(&w.res.errCnt, 1)
			}
			select {
			case w.res.errCh <- err:
			case <-w.done:
				return
			}
		}
	}()
}

func (w *worker) stop() {
	close(w.done)
}
