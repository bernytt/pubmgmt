package module

import (
	"errors"
	"strings"
)

type Shell struct {
	Environment []string
	Command     string
	Filter      bool
}

var dangerCommands = []string{"rm", "dd", "reboot", "halt", "init", "shutdown"}
var errDangerCommand = errors.New("Command is not allowed")

func contains(strs []string, substr string) bool {
	for _, str := range strs {
		if strings.Contains(substr, str) {
			return true
		}
	}
	return false
}

func (s *Shell) Build() (*ExecCommand, error) {
	if s.Filter {
		if contains(dangerCommands, s.Command) {
			return nil, errDangerCommand
		}
	}
	c := &ExecCommand{
		Environment: s.Environment,
		Command:     s.Command,
	}
	return c, nil
}

func (s *Shell) Name() string { return "shell" }

func init() {
	Modules["shell"] = func() Module { return &Shell{} }
}
