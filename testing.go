package paralleldl

import (
	"io/ioutil"
	"net/http"
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
