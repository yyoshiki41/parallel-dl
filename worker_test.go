package paralleldl

import "testing"

func TestErrCounts(t *testing.T) {
	c := createTestClient(t)
	d := newDispatcher(c, 1, 1)
	if actual := d.errCounts(); actual != 0 {
		t.Errorf("expected %d, but got %d", 0, actual)
	}
}
