package module

type Shell struct {
	Environment []string
	Command     string
}

func (s *Shell) Build() *ExecCommand {
	return &ExecCommand{
		Environment: s.Environment,
		Command:     s.Command,
	}
}

func (s *Shell) Name() string { return "shell" }

func init() {
	Modules["shell"] = func() Module { return &Shell{} }
}
