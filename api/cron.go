package pub

import (
	"fmt"
	"strings"
	"time"

	"github.com/fengxsong/pubmgmt/helper"
)

type Cron struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Type      string    `json:"type" binding:"required"`
	Cmd       []string  `json:"cmd,omitempty"`
	URL       string    `json:"url,omitempty"`
	Spec      string    `json:"spec" binding:"required"`
	Suspended bool      `json:"suspende"`
	Running   bool      `json:"running"`
	Times     int       `json:"times"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Err       error     `json:"-"`
	Output    string    `json:"-"`
	fails     int
}

func (*Cron) UniqueFields() []string {
	return []string{"ID", "Name"}
}

func (c *Cron) hasError(err error) {
	c.Err = err
	c.fails++
}

func (c *Cron) hasNotError() {
	c.Err, c.fails = nil, 0
}

func (c *Cron) Run() {
	if c.Times == 0 {
		c.Times = 1
	}
	c.Running = true
	defer func() { c.Running = false }()
	for i := 0; i < c.Times; i++ {
		switch strings.ToLower(c.Type) {
		case "url", "http":
			sess := helper.NewSession(nil, nil)
			resp, err := sess.Get(c.URL, nil, nil)
			if err != nil {
				c.hasError(err)
				continue
			}
			c.Output = fmt.Sprintf("Status code: %d, len(response): %d", resp.StatusCode, resp.ContentLength)
		case "cmd", "command", "shell":
			stdout, err := helper.Command(c.Cmd[0], c.Cmd[1:]...)
			if err != nil {
				c.hasError(err)
				continue
			}
			c.Output = stdout
		}
		c.hasNotError()
		break
	}
	c.Updated = time.Now()
}
