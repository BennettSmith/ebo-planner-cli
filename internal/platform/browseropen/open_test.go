package browseropen

import "testing"

func TestDefaultOpener_EmptyURLIsError(t *testing.T) {
	if err := (DefaultOpener{}).Open(""); err == nil {
		t.Fatalf("expected error")
	}
}
