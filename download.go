package paralleldl

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"sync/atomic"
)

var (
	errRequests int64
	quit        chan struct{}
	pool        chan *worker
	wg          sync.WaitGroup
)

func (c *Client) Do(list []string) (int64, error) {
	//ctx, ctxCancel := context.WithCancel(context.Background())
	//defer ctxCancel()

	maxQueues := len(list)
	maxWorkers := int(c.opt.MaxConcurrents)
	if maxWorkers == 0 {
		maxWorkers = len(list)
	}

	pool = make(chan *worker, maxWorkers)

	d := newDispatcher(c, maxQueues, maxWorkers)
	d.start()
	for _, v := range list {
		d.add(v)
	}

	d.wait()

	return errRequests, nil
}

type dispatcher struct {
	queue   chan string
	workers []*worker
}

func newDispatcher(c *Client, maxQueues, maxWorkers int) *dispatcher {
	d := &dispatcher{
		queue: make(chan string, maxQueues),
	}
	d.workers = make([]*worker, maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		w := worker{
			client: c,
			target: make(chan string),
			quit:   make(chan struct{}),
		}
		d.workers[i] = &w
	}
	return d
}

func (d *dispatcher) start() {
	for _, w := range d.workers {
		w.start()
	}
	go func() {
		for {
			select {
			case v := <-d.queue:
				worker := <-pool
				worker.target <- v
			case <-quit:
				d.stop()
			}
		}
	}()
}

func (d *dispatcher) stop() {
	for _, w := range d.workers {
		w.quit <- struct{}{}
	}
}

func (d *dispatcher) wait() {
	wg.Wait()
}

func (d *dispatcher) add(v string) {
	wg.Add(1)
	d.queue <- v
}

type worker struct {
	client *Client
	target chan string
	quit   chan struct{}
}

func (w *worker) start() {
	output := w.client.opt.Output
	maxAttempts := w.client.opt.MaxAttempts
	maxErrRequests := w.client.opt.MaxErrorRequests
	go func() {
		for {
			pool <- w
			select {
			case target := <-w.target:
				var req *http.Request
				var err error
				var attempts int64
				for {
					attempts++
					if maxAttempts != 0 && attempts > maxAttempts {
						// give up
						break
					}

					req, err = newRequest(context.Background(), target)
					if err != nil {
						break
					}

					b, retry, err := w.client.do(req)
					if err != nil {
						continue
					}
					if retry {
						err = errors.New("HTTP Status Code 5xx")
						continue
					}

					_, name := path.Split(target)
					err = createFile(path.Join(output, name), b)
					if err != nil {
						continue
					}

					// successful
					break
				}
				if err != nil {
					log.Println(err)
					atomic.AddInt64(&errRequests, 1)
					if maxErrRequests != 0 && errRequests >= maxErrRequests {
						close(quit)
					}
				}
				wg.Done()

			case <-w.quit:
				return
			}
		}
	}()
}

func newRequest(ctx context.Context, target string) (*http.Request, error) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	return req, nil
}

func (c *Client) do(req *http.Request) ([]byte, bool, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, true, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	return b, false, nil
}

func createFile(name string, body []byte) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	_, err = file.Write(body)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
