package paralleldl

import (
	"net/http/httptest"
	"testing"

	"github.com/fortytw2/leaktest"
)

func setupTestDownload(t *testing.T, opt *Options) (*Client, *httptest.Server, func()) {
	dir, removeDir := createTestTempDir(t)
	opt.Output = dir
	client, _ := New(opt)

	ts := runTestServer()

	return client, ts, func() { removeDir() }
}

func TestDownload(t *testing.T) {
	defer leaktest.Check(t)()

	opt := &Options{}
	client, ts, removeDir := setupTestDownload(t, opt)
	defer removeDir() // clean up
	defer ts.Close()

	lists := []string{
		ts.URL + "/ok1",
		ts.URL + "/ok2",
	}
	errCounts := client.Download(lists)
	if expected := int64(0); errCounts != expected {
		t.Errorf("expected %d, but got %d", expected, errCounts)
	}
}

func TestDownload_Error1(t *testing.T) {
	defer leaktest.Check(t)()

	opt := &Options{
		MaxAttempts: 1,
	}
	client, ts, removeDir := setupTestDownload(t, opt)
	defer removeDir() // clean up
	defer ts.Close()

	lists := []string{
		ts.URL + "/ok1",
		ts.URL + "/ok2",
		ts.URL + "/error",
	}
	errCounts := client.Download(lists)
	if expected := int64(1); errCounts != expected {
		t.Errorf("expected %d, but got %d", expected, errCounts)
	}
}

func TestDownload_Error2(t *testing.T) {
	defer leaktest.Check(t)()

	opt := &Options{
		MaxErrorRequests: 1,
		MaxAttempts:      2,
	}
	client, ts, removeDir := setupTestDownload(t, opt)
	defer removeDir() // clean up
	defer ts.Close()

	lists := []string{
		ts.URL + "/ok1",
		ts.URL + "/error",
		ts.URL + "/error",
	}
	errCounts := client.Download(lists)
	if expected := int64(2); errCounts != expected {
		t.Errorf("expected %d, but got %d", expected, errCounts)
	}
}

func TestDownload_Error3(t *testing.T) {
	defer leaktest.Check(t)()

	opt := &Options{
		MaxConcurrents:   1,
		MaxErrorRequests: 1,
		MaxAttempts:      1024,
	}
	client, ts, removeDir := setupTestDownload(t, opt)
	defer removeDir() // clean up
	defer ts.Close()

	lists := []string{
		ts.URL + "/error",
		ts.URL + "/error",
		ts.URL + "/error",
	}
	errCounts := client.Download(lists)
	if expected := int64(2); errCounts != expected {
		t.Errorf("expected %d, but got %d", expected, errCounts)
	}
}

func TestDownload_Error4(t *testing.T) {
	defer leaktest.Check(t)()

	opt := &Options{
		MaxConcurrents:   1,
		MaxErrorRequests: 1,
		MaxAttempts:      1024,
	}
	client, ts, removeDir := setupTestDownload(t, opt)
	defer removeDir() // clean up
	defer ts.Close()

	lists := []string{
		ts.URL + "/not-found",
		ts.URL + "/not-found",
	}
	errCounts := client.Download(lists)
	if expected := int64(2); errCounts != expected {
		t.Errorf("expected %d, but got %d", expected, errCounts)
	}
}
