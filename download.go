package paralleldl

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

// Download downloads the lists resources.
func (c *Client) Download(list []string) int64 {
	maxQueues := len(list)
	maxWorkers := int(c.opt.MaxConcurrents)
	if maxWorkers == 0 {
		maxWorkers = len(list)
	}

	d := newDispatcher(c, maxQueues, maxWorkers)
	d.start(list)
	d.wait()
	return d.errCounts()
}

func (c *Client) download(ctx context.Context, target string) error {
	req, err := newRequest(ctx, target)
	if err != nil {
		return err
	}

	maxAttempts := c.opt.MaxAttempts
	var (
		attempts int64
		b        []byte
		retry    bool
	)
	for {
		attempts++
		if maxAttempts != 0 && attempts > maxAttempts {
			// give up
			break
		}

		b, retry, err = c.do(req)
		if err != nil {
			if !retry {
				break
			}
			continue
		}
		if retry {
			err = errors.New("HTTP Status Code 5xx")
			continue
		}

		_, name := path.Split(target)
		err = createFile(path.Join(c.opt.Output, name), b)
		if err != nil {
			log.Print(err)
			continue
		}

		// successful
		break
	}
	return err
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

	statusCode := resp.StatusCode
	if 400 <= statusCode && statusCode < 500 {
		return nil, false, fmt.Errorf("Client Error: %d", resp.StatusCode)
	}

	if statusCode >= 500 {
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
