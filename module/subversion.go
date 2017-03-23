package module

type EventType int

const (
	INFO EventType = iota
	LOG
	CHECKOUT
	EXPORT
	SWITCH
	UPDATE
	REVERT
)

var SvnEventTypes = map[EventType]string{
	INFO:     "info",
	LOG:      "log",
	CHECKOUT: "checkout",
	EXPORT:   "export",
	SWITCH:   "switch",
	UPDATE:   "update",
	REVERT:   "revert",
}

type Subversion struct {
	Environment []string
	Dest        string    `json:"dest"`
	Repo        string    `json:"repo"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	SvnPath     string    `json:"svn_path"`
	Revision    string    `json:"revision"`
	Force       bool      `json:"force"`
	EventType   EventType `json:"event_type"`
}

func (s *Subversion) Name() string { return "subversion" }

func (s *Subversion) Build() *ExecCommand {
	if s.SvnPath == "" {
		s.SvnPath = "/usr/bin/svn"
	}
	c := &ExecCommand{
		Environment: s.Environment,
		Command:     s.SvnPath,
	}
	c.Arguments = []string{
		"--non-interactive",
		"--trust-server-cert",
		"--no-auth-cache",
	}
	if s.Username != "" {
		c.Arguments = append(c.Arguments, []string{"--username", s.Username}...)
	}
	if s.Password != "" {
		c.Arguments = append(c.Arguments, []string{"--password", s.Password}...)
	}

	if s.Revision == "" {
		s.Revision = "HEAD"
	}
	switch s.EventType {
	case INFO:
		c.Arguments = append(c.Arguments, []string{"info", s.Dest}...)
	case LOG:
		c.Arguments = append(c.Arguments, []string{"log", s.Dest, "-r", s.Revision}...)
	case CHECKOUT:
		c.Arguments = append(c.Arguments, []string{"checkout", "-r", s.Revision, s.Repo, s.Dest}...)
	case EXPORT:
		var force string
		if s.Force {
			force = "--force"
		}
		c.Arguments = append(c.Arguments, []string{"export", force, "-r", s.Revision, s.Repo, s.Dest}...)
	case SWITCH:
		c.Arguments = append(c.Arguments, []string{"switch", s.Repo, s.Dest}...)
	case UPDATE:
		c.Arguments = append(c.Arguments, []string{"update", "-r", s.Revision, s.Dest}...)
	case REVERT:
		c.Arguments = append(c.Arguments, []string{"revert", "-R", s.Dest}...)
	default:
	}
	return c
}

func init() {
	Modules["subversion"] = func() Module { return &Subversion{} }
}
