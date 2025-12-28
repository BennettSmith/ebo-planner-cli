package editmode

import (
	"os"
	"os/exec"
	"strings"
)

// ResolveEditor returns the editor command name and args per CLI spec:
// $EBO_EDITOR if set, else $EDITOR, else "vi".
func ResolveEditor() (string, []string) {
	raw := strings.TrimSpace(os.Getenv("EBO_EDITOR"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if raw == "" {
		return "vi", nil
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "vi", nil
	}
	return parts[0], parts[1:]
}

var execCommand = exec.Command

// EditFile opens the file in the resolved editor and blocks until it exits.
// It returns the file contents after editing.
func EditFile(path string) ([]byte, error) {
	editor, args := ResolveEditor()
	args = append(args, path)
	cmd := execCommand(editor, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// EditTemp starts from the given YAML template, lets the user edit, then returns the final buffer.
// The temp file uses an unknown extension so downstream parsing can try JSON then YAML.
func EditTemp(yamlTemplate string) ([]byte, error) {
	f, err := os.CreateTemp("", "ebo-edit-*.req")
	if err != nil {
		return nil, err
	}
	name := f.Name()
	_ = f.Close()
	defer func() { _ = os.Remove(name) }()

	if yamlTemplate == "" {
		yamlTemplate = "{}\n"
	}
	if !strings.HasSuffix(yamlTemplate, "\n") {
		yamlTemplate += "\n"
	}
	if err := os.WriteFile(name, []byte(yamlTemplate), 0o600); err != nil {
		return nil, err
	}

	return EditFile(name)
}
