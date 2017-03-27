package http

import (
	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/gin-gonic/gin.v1"
)

type Server struct {
	Flags         *pub.CliFlags
	Logger        logger
	CryptoService pub.CryptoService
	JWTService    pub.JWTService
	UserService   pub.UserService
	HostService   pub.HostService
	MailerService pub.MailerService
	TaskService   pub.TaskService
	ModuleService pub.ModuleService
}

func (s *Server) Start() error {
	app := gin.Default()
	if !*s.Flags.Debug {
		gin.SetMode("release")
	}
	jwt := &middleWareService{jwtService: s.JWTService, authDisabled: *s.Flags.NoAuth}
	jwtAuth := jwt.mwCheckAuthentication()
	jwtAdmin := jwt.mwCheckAdministratorRole()
	auth := &AuthHandler{Logger: s.Logger, CryptoService: s.CryptoService, JWTService: s.JWTService, UserService: s.UserService}
	user := &UserHandler{Logger: s.Logger, CryptoService: s.CryptoService, JWTService: s.JWTService, UserService: s.UserService}
	host := &HostHandler{Logger: s.Logger, HostService: s.HostService}
	mailer := newMailerHandler(s.UserService, s.MailerService, s.Flags)
	task := newTaskHandler(s.Logger, s.HostService, s.TaskService, s.Flags)
	modules := &ModuleHandler{Logger: s.Logger, ModuleService: s.ModuleService}
	app.PUT("/users", user.createUser)
	app.GET("/users", jwtAuth, jwtAdmin, user.getUsers)
	app.GET("/users/profile/:id", jwtAuth, user.getUserProfileByID)
	app.POST("/users/login", auth.signIn)
	app.POST("/users/profile/:id", jwtAuth, user.updateUserProfileByID)
	app.POST("/users/admin/init", user.initAdmin)
	app.DELETE("/users/profile/:id", jwtAuth, user.deleteUserByID)
	app.PUT("/hosts", jwtAuth, jwtAdmin, host.createHost)
	app.GET("/hosts", jwtAuth, host.getHosts)
	app.GET("/hosts/pk/:id", jwtAuth, host.getHostByID)
	app.POST("/hosts/pk/:id", jwtAuth, jwtAdmin, host.updateHostByID)
	app.DELETE("/hosts/pk/:id", jwtAuth, jwtAdmin, host.deleteHostByID)
	app.PUT("/hostgroups", jwtAuth, jwtAdmin, host.createHostgroup)
	app.GET("/hostgroups", jwtAuth, host.getHostgroups)
	app.GET("/hostgroups/pk/:id", jwtAuth, host.getHostgroupByID)
	app.DELETE("/hostgroups/pk/:id", jwtAuth, jwtAdmin, host.deleteHostgroupByID)
	app.PUT("/mailer", jwtAuth, mailer.createEmail)
	app.GET("/mailer/:uuid", mailer.getEmailDetail)
	app.PUT("/tasks", jwtAuth, task.createTask)
	app.GET("/tasks", jwtAuth, task.getTasks)
	app.GET("/tasks/detail/:id", jwtAuth, task.getTaskByID)
	app.POST("/tasks/detail/:id", jwtAuth, task.modifyTaskByID)
	app.GET("/tasks/events/:id", task.getTaskEventByID)
	app.GET("/tasks/active/:id", jwtAuth, jwtAdmin, task.activeTaskByID)
	app.PUT("/crons", jwtAuth, jwtAdmin, task.createCronJob)
	app.GET("/crons", jwtAuth, task.getCronJobs)
	app.GET("/crons/detail/:id", jwtAuth, task.getCronJobByID)
	app.POST("/crons/detail/:id", jwtAuth, jwtAdmin, task.modifyCronJobByID)
	app.DELETE("/crons/detail/:id", jwtAuth, jwtAdmin, task.deleteCronJobByID)
	app.PUT("/modules/svn", jwtAuth, jwtAdmin, modules.createSvnInfo)
	app.GET("/modules/svn", jwtAuth, modules.getSvnInfos)
	app.GET("/modules/svn/:id", jwtAuth, modules.getSvnByID)
	app.POST("/modules/svn/:id", jwtAuth, jwtAdmin, modules.updateSvnByID)
	app.DELETE("/modules/svn/:id", jwtAuth, jwtAdmin, modules.deleteSvnByID)
	go user.checkAdminExists()
	return app.Run(*s.Flags.Addr)
}
