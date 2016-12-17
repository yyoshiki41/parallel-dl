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

func (c *Client) Do(list []string, output string) (int64, error) {
	var (
		errCounts int64
		wg        sync.WaitGroup
	)

	sem := make(chan struct{}, c.opt.MaxConcurrents)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	for _, v := range list {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()

			var err error
			for i := 0; int64(i) < c.opt.MaxAttempts; i++ {
				if errCounts > c.opt.MaxErrorRequests {
					break
				}

				sem <- struct{}{}
				b, retry, err := c.download(ctx, target)
				<-sem
				if err != nil {
					continue
				}
				if retry {
					err = errors.New("StatusCode 5xx")
					continue
				}

				err = save(target, output, b)
				if err != nil {
					break
				}
			}
			if err != nil {
				atomic.AddInt64(&errCounts, 1)
				log.Println(err)
			}
		}(v)
	}
	wg.Wait()

	return errCounts, nil
}

func (c *Client) download(ctx context.Context, target string) ([]byte, bool, error) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, false, err
	}
	req = req.WithContext(ctx)

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

func save(target, output string, body []byte) error {
	_, fileName := path.Split(target)
	file, err := os.Create(path.Join(output, fileName))
	if err != nil {
		return err
	}

	// TODO
	_, err = file.Write(body)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
