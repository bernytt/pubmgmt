package cli

import (
	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

func ParseFlags() (*pub.CliFlags, error) {
	flags := &pub.CliFlags{
		Addr:       kingpin.Flag("bind", "Address and port to serve pubmgmt").Default(":8080").Short('p').String(),
		NoAuth:     kingpin.Flag("no-auth", "Disable authentication").Default("false").Bool(),
		SmtpServer: kingpin.Flag("smtp", "smtp server address and port").Default("smtp.qiye.163.com:25").Short('h').String(),
		Username:   kingpin.Flag("username", "username of mailer").Default("username@example.com").Short('u').String(),
		Password:   kingpin.Flag("password", "password of mailer").Default("s3cret").Short('P').String(),
		FromAlias:  kingpin.Flag("from", "from alias of mailer").Default("pubmgmt").String(),
		MaxRetry:   kingpin.Flag("retry", "max retry times").Default("3").Int(),
		QueueSize:  kingpin.Flag("coroutine", "sending mail or task queue size").Default("128").Short('c').Int(),
		Data:       kingpin.Flag("data", "Path to the folder where the data is stored").Default(".").Short('d').String(),
	}
	kingpin.Parse()
	return flags, nil
}
