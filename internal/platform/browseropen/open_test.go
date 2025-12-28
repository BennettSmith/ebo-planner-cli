package browseropen

import "testing"

func TestDefaultOpener_EmptyURLIsError(t *testing.T) {
	if err := (DefaultOpener{}).Open(""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDefaultOpener_NonEmptyURLDoesNotReturnEmptyURLError(t *testing.T) {
	err := (DefaultOpener{}).Open("http://example.invalid")
	if err != nil && err.Error() == "empty url" {
		t.Fatalf("unexpected empty-url error")
	}
}
