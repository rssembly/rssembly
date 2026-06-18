package repo

import "testing"

func TestErrNotFound(t *testing.T) {
	if ErrNotFound == nil {
		t.Fatal("ErrNotFound should not be nil")
	}
}
