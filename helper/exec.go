package helper

import (
	"os/exec"
	"strings"
)

type Cmd struct {
	*exec.Cmd
}

func (cmd *Cmd) Arguments(args ...string) *Cmd {
	cmd.Cmd.Args = append(cmd.Cmd.Args[0:1], args...)
	return cmd
}

func (cmd *Cmd) AppendArguments(args ...string) *Cmd {
	cmd.Cmd.Args = append(cmd.Cmd.Args, args...)
	return cmd
}

func (cmd *Cmd) ResetArguments() *Cmd {
	cmd.Args = cmd.Args[0:1]
	return cmd
}

func (cmd *Cmd) Directory(workingDir string) *Cmd {
	cmd.Cmd.Dir = workingDir
	return cmd
}

func CommandBuilder(command string, args ...string) *Cmd {
	return &Cmd{Cmd: exec.Command(command, args...)}
}

func Command(command string, a ...string) (output string, err error) {
	var out []byte
	if len(a) == 0 {
		args := strings.Split(command, " ")
		for _, arg := range args {
			if arg[0] == '-' {
				a = append(a, arg)
			}
		}
	}
	out, err = exec.Command(command, a...).Output()
	if err == nil {
		output = string(out)
	}
	return
}
