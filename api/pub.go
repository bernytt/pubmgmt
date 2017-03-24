package pub

import (
	"time"

	"github.com/fengxsong/pubmgmt/helper"
	"github.com/kataras/go-mailer"
)

type (
	DataStore interface {
		Open() error
		Close() error
	}

	Server interface {
		Start() error
	}

	Model interface {
		UniqueFields() []string
	}

	CryptoService interface {
		Hash(data string) (string, error)
		CompareHashAndData(hash string, data string) error
	}

	JWTService interface {
		GenerateToken(data *TokenData) (string, error)
		ParseAndVerifyToken(token string) (*TokenData, error)
	}

	UserService interface {
		User(ID uint64) (*User, error)
		UserByUsername(username string) (*User, error)
		UserByEmail(email string) (*User, error)
		UsersByRole(role UserRole) ([]User, error)
		Users() ([]User, error)
		UpdateUser(ID uint64, user *User) error
		CreateUser(user *User) error
		DeleteUser(ID uint64) error
	}

	HostService interface {
		Hostgroup(ID uint64) (*Hostgroup, error)
		Hostgroups() ([]Hostgroup, error)
		HostgroupByName(name string) (*Hostgroup, error)
		UpdateHostgroup(ID uint64, hostgroup *Hostgroup) error
		CreateHostgroup(hostgroup *Hostgroup) error
		DeleteHostgroup(ID uint64) error
		Host(ID uint64) (*Host, error)
		Hosts() ([]Host, error)
		HostByName(hostname string) (*Host, error)
		HostsByStatus(status bool) ([]Host, error)
		HostsByHostgroupID(ID uint64) ([]Host, error)
		UpdateHost(ID uint64, host *Host) error
		CreateHost(host *Host) error
		DeleteHost(ID uint64) error
		NewHost(str string) *Host
	}

	MailerService interface {
		CreateEmail(email *Email) error
		EmailByUser(userId uint64) ([]Email, error)
		EmailByUUID(uuid string) (*Email, error)
	}

	TaskService interface {
		Task(ID uint64) (*Task, error)
		TasksSchedule() ([]Task, error)
		Tasks(reqApproval, unfinished bool) ([]Task, error)
		UpdateTask(ID uint64, task *Task) error
		CreateTask(task *Task) error
		DeleteTask(ID uint64) error
		Cron(ID uint64) (*Cron, error)
		CronByName(name string) (*Cron, error)
		Crons() ([]Cron, error)
		UpdateCron(ID uint64, cron *Cron) error
		CreateCron(cron *Cron) error
		DeleteCron(ID uint64) error
	}
)

const (
	_ UserRole = iota
	AdministratorRole
	StandardUserRole
)

type (
	CliFlags struct {
		Addr       *string
		NoAuth     *bool
		SmtpServer *string
		Username   *string
		Password   *string
		FromAlias  *string
		MaxRetry   *int
		QueueSize  *int
		Data       *string
		Debug      *bool
	}

	UserRole uint64

	User struct {
		ID       uint64   `json:"id"`
		Username string   `json:"username" binding:"required,alphanum"`
		Email    string   `json:"email" binding:"required,email"`
		Password string   `json:"password" binding:"required,min=6,max=32"`
		Role     UserRole `json:"role"`
		Avatar   string   `json:"avatar,omitempty"`
		IsActive bool     `json:"is_active"`
	}

	TokenData struct {
		ID       uint64
		Username string
		Role     UserRole
	}

	Hostgroup struct {
		ID      uint64 `json:"id"`
		Name    string `json:"name" binding:"required"`
		Comment string `json:"comment"`
	}

	Host struct {
		ID           uint64 `json:"id"`
		Hostname     string `json:"hostname" binding:"required"`
		Username     string `json:"username"`
		Port         string `json:"port"`
		Password     string `json:"password,omitempty"`
		HostgroupID  uint64 `json:"hostgroup_id"`
		IdentityFile string `json:"identity_file,omitempty"`
		Comment      string `json:"comment"`
		IsActive     bool   `json:"is_active"`
	}

	Email struct {
		ID         uint64         `json:"id"`
		FromUserID uint64         `json:"user_id"`
		Cfg        *mailer.Config `json:"config,omitempty"`
		Subject    string         `json:"subject" binding:"required"`
		Content    string         `json:"content" binding:"required"`
		Tos        string         `json:"tos" binding:"required"`
		Created    time.Time      `json:"created"`
		Done       time.Time      `json:"done"`
		UUID       string         `json:"uuid"`
		Err        string         `json:"error"`
		Sender     mailer.Service `json:"-"`
	}

	Task struct {
		ID               uint64     `json:"id"`
		Name             string     `json:"name"`
		RequiredUserID   uint64     `json:"user_id,omitempty"`
		PreScript        string     `json:"pre_script,omitempty"`
		Command          [][]string `json:"command"`
		PostScript       string     `json:"post_script,omitempty"`
		Created          time.Time  `json:"created"`
		Done             time.Time  `json:"done"`
		Spec             string     `json:"spec,omitempty"`
		UUID             string     `json:"uuid"`
		Comment          string     `json:"comment"`
		RequiredApproval bool       `json:"required_approval"`
		Suspended        bool       `json:"suspended"`
		Hosts            []string   `json:"hosts"`
	}
)

func (*User) UniqueFields() []string {
	return []string{"ID", "Username", "Email"}
}

func (*Hostgroup) UniqueFields() []string {
	return []string{"ID", "Name"}
}

func (*Host) UniqueFields() []string {
	return []string{"ID", "Hostname"}
}

func (*Email) UniqueFields() []string {
	return []string{"ID", "UUID"}
}

func (*Task) UniqueFields() []string {
	return []string{"ID", "Name", "UUID"}
}

func NewEmail() Email {
	return Email{
		Created: time.Now(),
		UUID:    helper.NewUUID().String(),
	}
}

func (t *Task) Strings() [][]string {
	var commands [][]string
	if t.PreScript != "" {
		commands = append(commands, []string{"PreScript", t.PreScript})
	}
	commands = append(commands, t.Command...)
	if t.PostScript != "" {
		commands = append(commands, []string{"PostScript", t.PostScript})
	}
	return commands
}
