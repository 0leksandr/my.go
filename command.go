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

	if onStdout != nil {
		stdout, err := command.StdoutPipe()
		PanicIf(err)
		//defer func() { Must(stdout.Close()) }()
		stdoutScanner := bufio.NewScanner(stdout)
		go func() {
			for stdoutScanner.Scan() { onStdout(stdoutScanner.Text()) }
		}()
	}

	if onStderr != nil {
		stderr, err := command.StderrPipe()
		PanicIf(err)
		// &fs.PathError{Op:"close", Path:"|0", Err:(*errors.errorString)(0xc0001121c0)}
		// close |0: file already closed
		//defer func() { Must(stderr.Close()) }()
		stderrScanner := bufio.NewScanner(stderr)
		go func() {
			for stderrScanner.Scan() { onStderr(stderrScanner.Text()) }
		}()
	}

	return command.Run() == nil
}
func WriteToStderr(text string) {
	_, err := fmt.Fprintln(os.Stderr, text)
	PanicIf(err)
}
