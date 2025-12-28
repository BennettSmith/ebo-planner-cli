package browseropen

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Opener interface {
	Open(url string) error
}

type DefaultOpener struct{}

var (
	goos        = runtime.GOOS
	execCommand = exec.Command
)

func (DefaultOpener) Open(url string) error {
	if url == "" {
		return fmt.Errorf("empty url")
	}
	var cmd *exec.Cmd
	switch goos {
	case "darwin":
		cmd = execCommand("open", url)
	case "windows":
		cmd = execCommand("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = execCommand("xdg-open", url)
	}
	return cmd.Start()
}
