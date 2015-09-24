package command

import (
	"os/exec"
	"bytes"
)

func GetCommandResponse(binary string, cmd ...string) (bytes.Buffer, bytes.Buffer, error) {
	command := exec.Command(binary, cmd...)
	var stdout bytes.Buffer
	command.Stdout = &stdout
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err := command.Run()
	return stdout, stderr, err
}

func RunCommand(binary string, cmd ...string) {
	command := exec.Command(binary, cmd...)
	command.Run()
}