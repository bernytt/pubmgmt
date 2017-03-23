package ssh

import (
	"fmt"
)

type ExitError struct {
	Inner error
}

func (e *ExitError) Error() string {
	if e.Inner == nil {
		return "error"
	}
	return e.Inner.Error()
}

type Result struct {
	Stage  string `json:"Stage"`
	Err    error  `json:"Err,omitempty"`
	RC     int    `json:"ReturnCode,omitempty"`
	Stdout string `json:"Stdout,omitempty"`
	Stderr string `json:"Stderr,omitempty"`
}

func (r *Result) String() string {
	if r.Err != nil {
		return fmt.Sprintf("Stage: %s, return code(%d), error: %s, stderr: %s", r.Stage, r.RC, r.Err, r.Stderr)
	}
	return fmt.Sprintf("Stage: complete, all stdout: %s", r.Stdout)
}
