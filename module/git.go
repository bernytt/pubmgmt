package module

type Git struct {
	Environment []string
	Dest        string    `json:"dest"`
	Repo        string    `json:"repo"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	GitPath     string    `json:"git_path"`
	Revision    string    `json:"revision"`
	Force       bool      `json:"force"`
	EventType   EventType `json:"event_type"`
}

func (g *Git) Build() (*ExecCommand, error) {
	c := &ExecCommand{
		Environment: g.Environment,
		Command:     g.GitPath,
	}
	return c, nil
}

func (g *Git) Name() string { return "git" }

func init() {
	Modules["git"] = func() Module { return &Git{} }
}
