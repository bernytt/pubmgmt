package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fengxsong/pubmgmt/api"
	"github.com/kataras/go-mailer"
	"gopkg.in/gin-gonic/gin.v1"
)

type MailerHandler struct {
	Logger        logger
	UserService   pub.UserService
	MailerService pub.MailerService
	defaultSender mailer.Service
	incoming      chan *pub.Email
	processing    chan *pub.Email
	mq            chan struct{}
	retry         int
}

func newMailerHandler(u pub.UserService, m pub.MailerService, flags *pub.CliFlags) *MailerHandler {
	smtpServer := strings.Split(*flags.SmtpServer, ":")
	var smtpPort int
	if len(smtpServer) != 2 {
		smtpPort = 25
	} else {
		smtpPort, _ = strconv.Atoi(smtpServer[1])
	}

	mailer := &MailerHandler{
		UserService:   u,
		MailerService: m,
		defaultSender: mailer.New(
			mailer.Config{
				Host:      smtpServer[0],
				Port:      smtpPort,
				Username:  *flags.Username,
				Password:  *flags.Password,
				FromAlias: *flags.FromAlias,
			}),
		incoming:   make(chan *pub.Email, 1<<8),
		processing: make(chan *pub.Email, 1<<8),
		mq:         make(chan struct{}, *flags.QueueSize),
		retry:      *flags.MaxRetry,
	}
	go mailer.pickup()
	go mailer.sendQueue()
	return mailer
}

// pickup email from channel m.incoming, and then send it to m.processing
func (m *MailerHandler) pickup() {
	for {
		select {
		case email := <-m.incoming:
			if email.Cfg != nil {
				email.Sender = mailer.New(*email.Cfg)
				email.Cfg = nil
			} else {
				email.Sender = m.defaultSender
			}
			m.processing <- email
		}
	}
}

// get email from channel m.processing and deliver it to Email.Tos
func (m *MailerHandler) sendQueue() {
	for {
		select {
		case email := <-m.processing:
			m.mq <- struct{}{}
			go func(e *pub.Email, mq chan struct{}) {
				tos := strings.Split(e.Tos, ",")
				var err error
				for i := 0; i < m.retry; i++ {
					if err = e.Sender.Send(e.Subject, e.Content, tos...); err == nil {
						break
					}
				}
				if err != nil {
					e.Err = err.Error()
				} else {
					e.Done = time.Now()
				}
				m.MailerService.CreateEmail(e)
				<-mq
			}(email, m.mq)
		}
	}
}

// url: /mailer  method: PUT  body: pub.Email
// receive email from request and send it to channel for preparing.
func (m *MailerHandler) createEmail(ctx *gin.Context) {
	var req = pub.NewEmail()
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, ErrInvalidJSON, http.StatusBadGateway, nil)
		return
	}
	tokenData, err := extractTokenDataFromRequestContext(ctx)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, nil)
		return
	}
	req.FromUserID = tokenData.ID
	m.incoming <- &req
	ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: fmt.Sprintf("Check `%s/mailer/%s` for detail later", ctx.Request.Host, req.UUID)})
}

// url: /mailer/:uuid  method: GET
// get mail result
func (m *MailerHandler) getEmailDetail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if email, err := m.MailerService.EmailByUUID(uuid); err == pub.ErrEmailNotFound {
		Error(ctx, err, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, nil)
	} else {
		ctx.JSON(http.StatusOK, email)
	}
}
