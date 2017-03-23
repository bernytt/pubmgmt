package ssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"

	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/module"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	sshRetryInterval = 3
	defaultTimeout   = 60
)

type Client struct {
	Host           *pub.Host
	Stdout         bytes.Buffer
	Stderr         bytes.Buffer
	ConnectRetries int
	Timeout        int
	cli            *ssh.Client
}

func (s *Client) getSSHKey(identityFile string) (key ssh.Signer, err error) {
	buf, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, err
	}
	key, err = ssh.ParsePrivateKey(buf)
	return key, err
}

func (s *Client) getSSHAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if s.Host.Password != "" {
		methods = append(methods, ssh.Password(s.Host.Password))
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		defer sshAgent.Close()
	}

	if s.Host.IdentityFile == "" {
		s.Host.IdentityFile = os.Getenv("HOME") + "/.ssh/id_rsa"
	}
	key, err := s.getSSHKey(s.Host.IdentityFile)
	if err != nil && s.Host.Password == "" {
		return nil, err
	}
	methods = append(methods, ssh.PublicKeys(key))

	return methods, nil
}

func (s *Client) Connect() error {
	if s.Timeout == 0 {
		s.Timeout = defaultTimeout
	}
	methods, err := s.getSSHAuthMethods()
	if err != nil {
		return err
	}
	config := &ssh.ClientConfig{
		User: s.Host.Username,
		Auth: methods,
	}
	connectRetries := s.ConnectRetries
	if connectRetries == 0 {
		connectRetries = 3
	}

	var finalError error
	for i := 0; i < connectRetries; i++ {
		client, err := ssh.Dial("tcp", s.Host.Hostname+":"+s.Host.Port, config)
		if err == nil {
			s.cli = client
			return nil
		}
		time.Sleep(sshRetryInterval * time.Second)
		finalError = err
	}
	return finalError
}

func (s *Client) newSession() (*ssh.Session, error) {
	if s.cli == nil {
		return nil, fmt.Errorf("Not connected")
	}
	session, err := s.cli.NewSession()
	if err != nil {
		return nil, err
	}

	session.Stdout = &s.Stdout
	session.Stderr = &s.Stderr
	return session, nil
}

func (s *Client) exec(session *ssh.Session, c []string, resultChan chan *Result, block chan struct{}) {
	defer session.Close()
	var (
		rc  int
		err error
	)
	err = session.Run(c[1])
	if err != nil {
		if err, ok := err.(*ssh.ExitError); ok {
			rc = err.Waitmsg.ExitStatus()
		}
	}
	resultChan <- &Result{c[0], err, rc, s.Stdout.String(), s.Stderr.String()}
	block <- struct{}{}
}

func (s *Client) Run(cmd module.Command) (result *Result) {
	block := make(chan struct{}, 1)
	for _, c := range cmd.Strings() {
		session, err := s.newSession()
		if err != nil {
			result = &Result{Err: err}
			return
		}
		resultChan := make(chan *Result)
		go s.exec(session, c, resultChan, block)
		select {
		case result = <-resultChan:
			<-block
			if result.Err != nil {
				return
			}
		case <-time.After(time.Duration(s.Timeout) * time.Second):
			result = &Result{Stage: "Runtime", Err: fmt.Errorf("Execution timed out")}
			return
		}
	}
	return
}

func (s *Client) Scp(filePath, destPath string) error {
	session, err := s.newSession()
	if err != nil {
		return err
	}
	defer session.Close()
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fstat, err := f.Stat()
	if err != nil {
		return err
	}
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%#o %d %s\n", fstat.Mode().Perm(), fstat.Size(), path.Base(filePath))
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()
	command := fmt.Sprintf("scp -qrt %s", destPath)
	if err := session.Run(command); err != nil {
		return err
	}
	return nil
}

func (s *Client) Cleanup() {
	if s.cli != nil {
		s.cli.Close()
	}
}
