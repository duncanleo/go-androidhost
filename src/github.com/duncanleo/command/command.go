package command

import (
	"os/exec"
	"bytes"
)

func GetCommandResponse(binary string, cmd ...string) (string, error) {
	command := exec.Command(binary, cmd...)
	var stdout bytes.Buffer
	command.Stdout = &stdout
	err := command.Run()
	return stdout.String(), err
}

func RunCommand(binary string, cmd ...string) {
	command := exec.Command(binary, cmd...)
	command.Run()
}