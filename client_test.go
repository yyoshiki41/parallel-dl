package paralleldl

import (
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	_, err := New(&Options{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNew_EmptyHTTPClient(t *testing.T) {
	var c *http.Client

	SetHTTPClient(c)
	defer teardownHTTPClient()

	_, err := New(&Options{})
	if err == nil {
		t.Error("Should detect that HTTPClient is nil.")
	}
}
