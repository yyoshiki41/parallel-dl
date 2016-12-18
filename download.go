package paralleldl

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
)

var (
	wg         sync.WaitGroup
	errChannel chan error
	workerPool chan *worker
)

// Download downloads the lists resources.
func (c *Client) Download(list []string) (int64, error) {
	maxQueues := len(list)
	maxWorkers := int(c.opt.MaxConcurrents)
	if maxWorkers == 0 {
		maxWorkers = len(list)
	}

	errChannel = make(chan error, maxQueues)
	workerPool = make(chan *worker, maxWorkers)

	d := newDispatcher(c, maxQueues, maxWorkers)
	d.start()
	for _, v := range list {
		d.add(v)
	}

	d.wait()
	return int64(len(errChannel)), nil
}

type dispatcher struct {
	jobQueue chan string
	workers  []*worker
}

func newDispatcher(c *Client, maxQueues, maxWorkers int) *dispatcher {
	d := &dispatcher{
		jobQueue: make(chan string, maxQueues),
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
				worker := <-workerPool
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
	client *Client
	target chan string
	quit   chan struct{}
}

func (w *worker) start(ctx context.Context) {
	output := w.client.opt.Output
	maxAttempts := w.client.opt.MaxAttempts
	go func() {
		for {
			workerPool <- w
			select {
			case target := <-w.target:
				var req *http.Request
				var err error
				req, err = newRequest(ctx, target)
				if err != nil {
					errChannel <- err
					wg.Done()
				}

				var attempts int64
				for {
					attempts++
					if maxAttempts != 0 && attempts > maxAttempts {
						// give up
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
