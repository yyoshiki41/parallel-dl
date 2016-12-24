package paralleldl

import (
	"testing"

	"github.com/fortytw2/leaktest"
)

func TestDownload(t *testing.T) {
	defer leaktest.Check(t)()

	dir, removeDir := createTestTempDir(t)
	defer removeDir() // clean up

	ts := runTestServer()
	defer ts.Close()

	client, err := New(&Options{Output: dir})
	if err != nil {
		t.Fatalf("Failed to construct client: %s", err)
	}

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

	dir, removeDir := createTestTempDir(t)
	defer removeDir() // clean up

	ts := runTestServer()
	defer ts.Close()

	opt := &Options{
		Output:      dir,
		MaxAttempts: 1,
	}
	client, err := New(opt)
	if err != nil {
		t.Fatalf("Failed to construct client: %s", err)
	}

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

	dir, removeDir := createTestTempDir(t)
	defer removeDir() // clean up

	ts := runTestServer()
	defer ts.Close()

	opt := &Options{
		Output:           dir,
		MaxErrorRequests: 1,
		MaxAttempts:      4,
	}
	client, err := New(opt)
	if err != nil {
		t.Fatalf("Failed to construct client: %s", err)
	}

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

	dir, removeDir := createTestTempDir(t)
	defer removeDir() // clean up

	ts := runTestServer()
	defer ts.Close()

	opt := &Options{
		Output:           dir,
		MaxConcurrents:   1,
		MaxErrorRequests: 1,
		MaxAttempts:      1024,
	}
	client, err := New(opt)
	if err != nil {
		t.Fatalf("Failed to construct client: %s", err)
	}

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

	dir, removeDir := createTestTempDir(t)
	defer removeDir() // clean up

	ts := runTestServer()
	defer ts.Close()

	opt := &Options{
		Output:           dir,
		MaxConcurrents:   1,
		MaxErrorRequests: 1,
		MaxAttempts:      1024,
	}
	client, err := New(opt)
	if err != nil {
		t.Fatalf("Failed to construct client: %s", err)
	}

	lists := []string{
		ts.URL + "/not-found",
		ts.URL + "/not-found",
	}
	errCounts := client.Download(lists)
	if expected := int64(2); errCounts != expected {
		t.Errorf("expected %d, but got %d", expected, errCounts)
	}
}
