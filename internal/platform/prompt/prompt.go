package prompt

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrAborted indicates the prompt session was aborted (e.g., SIGINT).
var ErrAborted = errors.New("prompt aborted")

type Editor func(template string) ([]byte, error)

type Prompter struct {
	In   io.Reader
	Out  io.Writer // prompt output (stderr by convention)
	Edit Editor    // used for multi-line fields (Enter to open editor)

	br *bufio.Reader
}

func New(in io.Reader, out io.Writer, edit Editor) *Prompter {
	return &Prompter{In: in, Out: out, Edit: edit}
}

func (p *Prompter) reader() *bufio.Reader {
	if p.br != nil {
		return p.br
	}
	in := p.In
	if in == nil {
		in = strings.NewReader("")
	}
	p.br = bufio.NewReader(in)
	return p.br
}

func (p *Prompter) PromptRequiredString(ctx context.Context, label string) (string, error) {
	for {
		s, err := p.PromptOptionalString(ctx, label)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(s) == "" {
			_, _ = fmt.Fprintf(p.Out, "%s is required.\n", label)
			continue
		}
		return s, nil
	}
}

func (p *Prompter) PromptOptionalString(ctx context.Context, label string) (string, error) {
	_, _ = fmt.Fprintf(p.Out, "%s: ", label)
	line, err := readLineCtx(ctx, p.reader())
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// PromptMultilineOrInline prompts for a potentially multi-line field.
// If the user enters a non-empty line, it's used as a single-line value.
// If the user presses Enter on an empty line, the editor is launched.
func (p *Prompter) PromptMultilineOrInline(ctx context.Context, label string, editorTemplate string) (string, bool, error) {
	_, _ = fmt.Fprintf(p.Out, "Enter %s (press Enter to open editor, or type inline for single line): ", label)
	line, err := readLineCtx(ctx, p.reader())
	if err != nil {
		return "", false, err
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) != "" {
		return line, false, nil
	}
	if p.Edit == nil {
		return "", false, fmt.Errorf("no editor configured")
	}
	b, err := p.Edit(editorTemplate)
	if err != nil {
		return "", false, err
	}
	return string(b), true, nil
}

func (p *Prompter) PromptYesNo(ctx context.Context, label string, defaultNo bool) (bool, error) {
	suffix := "(y/N)"
	if !defaultNo {
		suffix = "(Y/n)"
	}
	_, _ = fmt.Fprintf(p.Out, "%s %s: ", label, suffix)
	line, err := readLineCtx(ctx, p.reader())
	if err != nil {
		return false, err
	}
	v := strings.ToLower(strings.TrimSpace(strings.TrimRight(line, "\r\n")))
	if v == "" {
		return !defaultNo, nil
	}
	return v == "y" || v == "yes", nil
}

func (p *Prompter) PromptStringList(ctx context.Context, label string) ([]string, error) {
	var out []string
	for i := 1; ; i++ {
		_, _ = fmt.Fprintf(p.Out, "Enter %s #%d (or press Enter to finish): ", label, i)
		line, err := readLineCtx(ctx, p.reader())
		if err != nil {
			return nil, err
		}
		v := strings.TrimSpace(strings.TrimRight(line, "\r\n"))
		if v == "" {
			break
		}
		out = append(out, v)
	}
	return out, nil
}

func readLineCtx(ctx context.Context, r *bufio.Reader) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	// Special case: allow immediate abort in tests by canceling ctx.
	select {
	case <-ctx.Done():
		return "", ErrAborted
	default:
	}

	type res struct {
		s   string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		s, err := r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			ch <- res{"", err}
			return
		}
		// If EOF without newline, still return the partial.
		ch <- res{s, nil}
	}()

	select {
	case <-ctx.Done():
		return "", ErrAborted
	case r := <-ch:
		// Treat literal ^C (ETX) as abort as well (useful for deterministic tests).
		if strings.Contains(r.s, "\x03") {
			return "", ErrAborted
		}
		return r.s, r.err
	}
}
