package paralleldl

import (
	"bytes"
	"context"
	"path"
	"testing"
)

func TestDownload(t *testing.T) {
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
	errCounts, err := client.Download(lists)
	if err != nil {
		t.Error(err)
	}
	if errCounts != 0 {
		t.Errorf("expected %d, but got %d", 0, errCounts)
	}
}

func TestNewRequest(t *testing.T) {
	_, err := newRequest(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
}

func TestDo(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()

	cases := []struct {
		target   string
		expected []byte
		retry    bool
	}{
		{target: "/ok1", expected: []byte("OK1\n"), retry: false},
		{target: "/moved-permanently", expected: []byte("OK1\n"), retry: false},
		{target: "/bad-request", expected: []byte("Bad Request\n"), retry: false},
		{target: "/not-found", expected: []byte("404 page not found\n"), retry: false},
		{target: "/error", expected: []byte(""), retry: true},
	}

	client := createTestClient(t)
	for _, c := range cases {
		req, _ := newRequest(context.Background(), ts.URL+c.target)
		b, retry, err := client.do(req)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(b, c.expected) {
			t.Errorf("expected %s, but got %s: %s", string(c.expected), string(b), c.target)
		}
		if retry != c.retry {
			t.Errorf("expected %v, but got %v: %s", c.retry, retry, c.target)
		}
	}
}

func TestCreateFile(t *testing.T) {
	dir, removeDir := createTestTempDir(t)
	defer removeDir() // clean up

	err := createFile(path.Join(dir, "test.txt"), []byte("tests"))
	if err != nil {
		t.Error(err)
	}
}