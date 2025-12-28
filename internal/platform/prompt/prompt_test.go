package prompt

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPromptRequiredString_RepromptsUntilNonEmpty(t *testing.T) {
	in := strings.NewReader("\n\nok\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)

	v, err := p.PromptRequiredString(context.Background(), "Name")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if v != "ok" {
		t.Fatalf("got %q", v)
	}
}

func TestReadLineCtx_AbortedByCtrlCChar(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("\x03\n"))
	_, err := readLineCtx(context.Background(), in)
	if err != ErrAborted {
		t.Fatalf("expected ErrAborted, got %v", err)
	}
}

func TestPromptYesNo_DefaultNo(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)
	v, err := p.PromptYesNo(context.Background(), "Q", true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if v != false {
		t.Fatalf("got %v", v)
	}
}

func TestPromptYesNo_Yes(t *testing.T) {
	in := strings.NewReader("yes\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)
	v, err := p.PromptYesNo(context.Background(), "Q", true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if v != true {
		t.Fatalf("got %v", v)
	}
}

func TestPromptStringList_ReadsUntilBlank(t *testing.T) {
	in := strings.NewReader("a\nb\n\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)
	vs, err := p.PromptStringList(context.Background(), "artifactIds")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(vs) != 2 || vs[0] != "a" || vs[1] != "b" {
		t.Fatalf("got %#v", vs)
	}
}

func TestPromptMultilineOrInline_Inline(t *testing.T) {
	in := strings.NewReader("hi\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)
	s, usedEditor, err := p.PromptMultilineOrInline(context.Background(), "description", "ignored")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if usedEditor {
		t.Fatalf("expected inline path")
	}
	if s != "hi" {
		t.Fatalf("got %q", s)
	}
}

func TestPromptMultilineOrInline_Editor(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	p := New(in, out, func(template string) ([]byte, error) {
		_ = template
		return []byte("edited\n"), nil
	})
	s, usedEditor, err := p.PromptMultilineOrInline(context.Background(), "description", "tmpl")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !usedEditor {
		t.Fatalf("expected editor path")
	}
	if s != "edited\n" {
		t.Fatalf("got %q", s)
	}
}

func TestReadLineCtx_AbortedByCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := bufio.NewReader(strings.NewReader("x\n"))
	_, err := readLineCtx(ctx, in)
	if err != ErrAborted {
		t.Fatalf("expected ErrAborted, got %v", err)
	}
}

func TestPromptMultilineOrInline_Blank_NoEditorIsError(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)
	_, _, err := p.PromptMultilineOrInline(context.Background(), "description", "tmpl")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadLineCtx_EOFWithoutNewline_ReturnsPartial(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("x"))
	s, err := readLineCtx(context.Background(), in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s != "x" {
		t.Fatalf("got %q", s)
	}
}

func TestPromptYesNo_DefaultYes(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	p := New(in, out, nil)
	v, err := p.PromptYesNo(context.Background(), "Q", false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if v != true {
		t.Fatalf("got %v", v)
	}
}
