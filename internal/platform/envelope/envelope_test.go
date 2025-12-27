package envelope

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSON_SuccessEnvelope(t *testing.T) {
	buf := &bytes.Buffer{}
	env := Envelope{
		Data: map[string]any{"ok": true},
		Meta: Meta{APIURL: "http://x", Profile: "default"},
	}
	if err := WriteJSON(buf, env); err != nil {
		t.Fatalf("write: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["meta"] == nil {
		t.Fatalf("expected meta")
	}
	if got["data"] == nil {
		t.Fatalf("expected data")
	}
	if _, ok := got["error"]; ok {
		t.Fatalf("did not expect error")
	}
}

func TestWriteJSON_ErrorEnvelope(t *testing.T) {
	buf := &bytes.Buffer{}
	env := Envelope{
		Meta:  Meta{Profile: "p"},
		Error: &ErrorBody{Code: "SOME_ERROR", Message: "nope"},
	}
	if err := WriteJSON(buf, env); err != nil {
		t.Fatalf("write: %v", err)
	}

	var got struct {
		Meta  Meta      `json:"meta"`
		Error ErrorBody `json:"error"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Error.Message != "nope" {
		t.Fatalf("message: got %q", got.Error.Message)
	}
}
