package module

import (
	"bytes"
	"strings"

	"github.com/fengxsong/pubmgmt/helper"
)

var Modules = map[string]func() Module{}

type Module interface {
	Build() *ExecCommand
	Name() string
}

func GetModules() []string {
	var modules []string
	for k, _ := range Modules {
		modules = append(modules, k)
	}
	return modules
}

func NewExecCommand(m Module) *ExecCommand {
	return m.Build()
}

type Command interface {
	Strings() [][]string
}

type ExecCommand struct {
	Environment []string         `json:"environment"`
	Command     string           `json:"command"`
	Arguments   []string         `json:"arguments"`
	Abort       chan interface{} `json:"-"`
}

func (c *ExecCommand) buildCommand() string {
	arguments := []string{c.Command}
	for _, part := range c.Arguments {
		arguments = append(arguments, helper.ShellExcape(part))
	}
	return strings.Join(arguments, " ")
}

func (c *ExecCommand) Strings() [][]string {
	var buf bytes.Buffer
	for _, v := range c.Environment {
		buf.WriteString("export " + helper.ShellExcape(v) + "\n")
	}
	buf.WriteString(c.buildCommand())
	return [][]string{{"Command", buf.String()}}
}
