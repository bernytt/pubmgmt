package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/api/bolt"
	"github.com/fengxsong/pubmgmt/api/cli"
	"github.com/fengxsong/pubmgmt/api/crypto"
	"github.com/fengxsong/pubmgmt/api/http"
	"github.com/fengxsong/pubmgmt/api/jwt"
)

func initStore(dataStorePath string) *bolt.Store {
	store, err := bolt.NewStore(dataStorePath)
	if err != nil {
		log.Fatalln(err)
	}
	err = store.Open()
	if err != nil {
		log.Fatalln(err)
	}
	return store
}

func initJWTService(authenticationEnabled bool) pub.JWTService {
	if authenticationEnabled {
		jwtService, err := jwt.NewService()
		if err != nil {
			log.Fatalln(err)
		}
		return jwtService
	}
	return nil
}

func initCryptoService() pub.CryptoService {
	return &crypto.Service{}
}

func main() {
	flags, _ := cli.ParseFlags()
	store := initStore(*flags.Data)

	server := http.Server{
		Flags:         flags,
		Logger:        log.New(),
		CryptoService: initCryptoService(),
		JWTService:    initJWTService(true),
		UserService:   store.UserService,
		HostService:   store.HostService,
		MailerService: store.MailerService,
		TaskService:   store.TaskService,
		ModuleService: store.ModuleService,
	}
	err := server.Start()
	if err != nil {
		log.Fatalln(err)
	}
}
