package shellexec

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ExecUnix(cmd string) (*string, error) {
	return RunExec(exec.Command("sh", "-c", cmd))
}

func ExecWinshell(cmd string) (*string, error) {
	return RunExec(exec.Command("cmd.exe", "/C", cmd))
}

func RunExec(cmd *exec.Cmd) (*string, error) {
	// Capture output
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	// Get output result
	output := out.String()

	// If there is an error, return error information
	if err != nil {
		errorOutput := stderr.String()
		if errorOutput != "" {
			output += "\n" + errorOutput
		}
		return nil, fmt.Errorf("%v:%v", output, err)
	}

	return &output, nil
}
