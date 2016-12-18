package paralleldl

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var (
	errChannel chan error
)

// Download downloads the lists resources.
func (c *Client) Download(list []string) (int64, error) {
	maxQueues := len(list)
	maxWorkers := int(c.opt.MaxConcurrents)
	if maxWorkers == 0 {
		maxWorkers = len(list)
	}

	errChannel = make(chan error, maxQueues)

	d := newDispatcher(c, maxQueues, maxWorkers)
	d.start()

	for _, v := range list {
		d.add(v)
	}
	d.wait()
	return int64(len(errChannel)), nil
}

func (c *Client) download(ctx context.Context, target string) error {
	var (
		req *http.Request
		err error
	)

	req, err = newRequest(ctx, target)
	if err != nil {
		return err
	}

	var attempts int64
	maxAttempts := c.opt.MaxAttempts
	for {
		attempts++
		if maxAttempts != 0 && attempts > maxAttempts {
			// give up
			break
		}

		b, retry, err := c.do(req)
		if err != nil {
			continue
		}
		if retry {
			err = errors.New("HTTP Status Code 5xx")
			continue
		}

		_, name := path.Split(target)
		err = createFile(path.Join(c.opt.Output, name), b)
		if err != nil {
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
