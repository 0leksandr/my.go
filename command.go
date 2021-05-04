package my

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(dir string, cmd string, onStdout func(string), onStderr func(string)) bool {
	command := exec.Command("sh", "-c", cmd)
	command.Dir = dir

	stdout, err := command.StdoutPipe()
	PanicIf(err)
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			stdout := stdoutScanner.Text()
			if onStdout != nil { onStdout(stdout) }
		}
	}()

	stderr, err := command.StderrPipe()
	PanicIf(err)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			stderr := stderrScanner.Text()
			if onStderr != nil { onStderr(stderr) }
		}
	}()

	return command.Run() == nil
}
func WriteToStderr(text string) {
	_, err := fmt.Fprintln(os.Stderr, text)
	PanicIf(err)
}
