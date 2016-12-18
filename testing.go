package paralleldl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Should restore defaultHTTPClient if SetHTTPClient is called.
func teardownHTTPClient() {
	SetHTTPClient(&http.Client{})
}

func createTestTempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "test-parallel-dl")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}

	return dir, func() { os.RemoveAll(dir) }
}

func createTestClient(t *testing.T) *Client {
	c, err := New(&Options{})
	if err != nil {
		t.Fatalf("Failed to create test client: %s", err)
	}

	return c
}

func runTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/ok1",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "OK1")
		},
	)
	mux.HandleFunc(
		"/ok2",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "OK2")
		},
	)
	mux.HandleFunc(
		"/moved-permanently",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", r.URL.Host+"/ok1")
			w.WriteHeader(http.StatusMovedPermanently)
		},
	)
	mux.HandleFunc(
		"/bad-request",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Bad Request")
		},
	)
	mux.HandleFunc(
		"/not-found",
		func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		},
	)
	mux.HandleFunc(
		"/error",
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		},
	)

	return httptest.NewServer(mux)
}
