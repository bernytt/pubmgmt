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
	api := app.Group(*s.Flags.ApiPrefix)
	{
		api.PUT("/users", user.createUser)
		api.GET("/users", jwtAuth, jwtAdmin, user.getUsers)
		api.GET("/users/profile/:id", jwtAuth, user.getUserProfileByID)
		api.POST("/users/login", auth.signIn)
		api.POST("/users/profile/:id", jwtAuth, user.updateUserProfileByID)
		api.POST("/users/admin/init", user.initAdmin)
		api.DELETE("/users/profile/:id", jwtAuth, user.deleteUserByID)
		api.PUT("/hosts", jwtAuth, jwtAdmin, host.createHost)
		api.GET("/hosts", jwtAuth, host.getHosts)
		api.GET("/hosts/pk/:id", jwtAuth, host.getHostByID)
		api.POST("/hosts/pk/:id", jwtAuth, jwtAdmin, host.updateHostByID)
		api.DELETE("/hosts/pk/:id", jwtAuth, jwtAdmin, host.deleteHostByID)
		api.PUT("/hostgroups", jwtAuth, jwtAdmin, host.createHostgroup)
		api.GET("/hostgroups", jwtAuth, host.getHostgroups)
		api.GET("/hostgroups/pk/:id", jwtAuth, host.getHostgroupByID)
		api.DELETE("/hostgroups/pk/:id", jwtAuth, jwtAdmin, host.deleteHostgroupByID)
		api.PUT("/mailer", jwtAuth, mailer.createEmail)
		api.GET("/mailer/:uuid", mailer.getEmailDetail)
		api.PUT("/tasks", jwtAuth, task.createTask)
		api.GET("/tasks", jwtAuth, task.getTasks)
		api.GET("/tasks/detail/:id", jwtAuth, task.getTaskByID)
		api.POST("/tasks/detail/:id", jwtAuth, task.modifyTaskByID)
		api.GET("/tasks/events/:id", task.getTaskEventByID)
		api.GET("/tasks/active/:id", jwtAuth, jwtAdmin, task.activeTaskByID)
		api.PUT("/crons", jwtAuth, jwtAdmin, task.createCronJob)
		api.GET("/crons", jwtAuth, task.getCronJobs)
		api.GET("/crons/detail/:id", jwtAuth, task.getCronJobByID)
		api.POST("/crons/detail/:id", jwtAuth, jwtAdmin, task.modifyCronJobByID)
		api.DELETE("/crons/detail/:id", jwtAuth, jwtAdmin, task.deleteCronJobByID)
		api.PUT("/modules/svn", jwtAuth, jwtAdmin, modules.createSvnInfo)
		api.GET("/modules/svn", jwtAuth, modules.getSvnInfos)
		api.GET("/modules/svn/:id", jwtAuth, modules.getSvnByID)
		api.POST("/modules/svn/:id", jwtAuth, jwtAdmin, modules.updateSvnByID)
		api.DELETE("/modules/svn/:id", jwtAuth, jwtAdmin, modules.deleteSvnByID)
	}
	go user.checkAdminExists()
	return app.Run(*s.Flags.Addr)
}
