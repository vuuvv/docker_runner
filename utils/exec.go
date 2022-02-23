package utils

import (
	"bytes"
	"os/exec"
)

func RunCommand(name string, arg ...string) (string, error) {
	var output bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	return output.String(), err
}
