package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/helper"
	"github.com/fengxsong/pubmgmt/helper/ssh"
	"github.com/fengxsong/pubmgmt/module"
	"github.com/robfig/cron"
	"gopkg.in/gin-gonic/gin.v1"
)

type TaskHandler struct {
	Logger      logger
	HostService pub.HostService
	TaskService pub.TaskService
	incoming    chan *pub.Task
	scheduling  chan *pub.Task
	cache       *helper.Store
	cronPool    chan *pub.Cron
	cron        *cron.Cron
	events      chan *event
}

const (
	taskPrefix  = "task."
	eventPrefix = "event."
	cronPrefix  = "cron."
)

func newTaskHandler(l logger, h pub.HostService, t pub.TaskService, flags *pub.CliFlags) *TaskHandler {
	th := &TaskHandler{
		Logger:      l,
		HostService: h,
		TaskService: t,
		incoming:    make(chan *pub.Task, *flags.QueueSize),
		scheduling:  make(chan *pub.Task, *flags.QueueSize),
		cache:       helper.NewStore(),
		cronPool:    make(chan *pub.Cron, *flags.QueueSize),
		cron:        cron.New(),
		events:      make(chan *event, *flags.QueueSize*2),
	}
	go th.cron.Start()
	go th.initTasksFromStore()
	go th.process()
	go th.cacheResult()
	go th.initCrons()
	go th.runCron()
	return th
}

type event struct {
	task   *pub.Task
	Done   time.Time              `json:"done"`
	Result map[string]interface{} `json:"result"`
}

func (t *TaskHandler) initTasksFromStore() {
	// get tasks that require approval and unfinished.
	tasks, err := t.TaskService.Tasks(true, true)
	if err != nil && err != pub.ErrTaskSetEmpty {
		Infof(t.Logger, "Error when getting tasks require approval: %s", err)
		return
	}
	for _, task := range tasks {
		t.cache.Set(taskPrefix+task.UUID, &task, 0)
	}

	// send scheduling tasks into channel `scheduling`
	tasks, err = t.TaskService.TasksSchedule()
	if err != nil && err != pub.ErrTaskSetEmpty {
		Infof(t.Logger, "Error when getting scheduling tasks: %s", err)
		return
	}
	for _, task := range tasks {
		Infof(t.Logger, "sending scheduled task to scheduling channel: %s\n", task.Name)
		t.scheduling <- &task
	}
	// using the same solution as cron _(:з)∠)_
	go func() {
		for {
			select {
			case task := <-t.scheduling:
				if ti := t.cache.Get(taskPrefix + task.Name); ti != nil {
					task = ti.(*pub.Task)
					task.Suspended = true
					t.cache.Delete(taskPrefix + task.Name)
					if task.Suspended {
						break
					}
				}
				t.cache.Set(taskPrefix+task.Name, task, 0)
				t.cron.AddFunc(task.Spec, func() {
					if task.Suspended {
						return
					}
					t.incoming <- task
				})
			}
		}
	}()
}

// processing tasks in backgroud.
func (t *TaskHandler) process() {
	for {
		select {
		case task := <-t.incoming:
			Infof(t.Logger, "starting to exec task %s: %s\n", task.Name)
			evt := &event{
				task:   task,
				Result: make(map[string]interface{}),
			}
			for _, host := range task.Hosts {
				h, err := t.HostService.HostByName(host)
				if err != nil {
					evt.Result[host] = pub.ErrHostNotFound
					continue
				}
				if !h.IsActive {
					evt.Result[host] = pub.ErrHostInactive
					continue
				}
				cli := &ssh.Client{
					Host:   h,
					Stdout: bytes.Buffer{},
					Stderr: bytes.Buffer{},
				}

				err = cli.Connect()
				if err != nil {
					evt.Result[host] = err
				} else {
					evt.Result[host] = cli.Run(task).String()
				}
			}
			evt.Done = time.Now()
			t.events <- evt
		}
	}
}

// cache results
func (t *TaskHandler) cacheResult() {
	for {
		select {
		case evt := <-t.events:
			// is it neccessary to update task's `Done` field?
			task := evt.task
			task.Done = time.Now()
			t.TaskService.UpdateTask(task.ID, task)
			t.cache.Set(eventPrefix+task.UUID, evt, 0)
		}
	}
}

// url: /tasks  method: PUT  body: putTaskRequest
// using a ugly way to stop or restart the task.. see block 92-103
func (t *TaskHandler) createTask(ctx *gin.Context) {
	var req putTaskRequest
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, ErrInvalidJSON, http.StatusBadRequest, nil)
		return
	}
	if req.Spec != "" && len(strings.Split(req.Spec, " ")) != 6 {
		Error(ctx, pub.Error("Scheduled filed is not a valid crond format"), http.StatusBadRequest, nil)
		return
	}
	tokenData, err := extractTokenDataFromRequestContext(ctx)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
		return
	}
	m, ok := module.Modules[strings.ToLower(req.Module)]
	if !ok {
		Error(ctx, pub.Error(fmt.Sprintf("Module: %s not implement yet", req.Module)), http.StatusBadRequest, nil)
		return
	}
	reqModule := m()
	if err := json.Unmarshal(req.Data, reqModule); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	c, err := module.NewExecCommand(reqModule)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	task := &pub.Task{
		Name:             req.Name + time.Now().Format(".2006-01-02|15:04:05"),
		RequiredUserID:   tokenData.ID,
		Command:          c.Strings(),
		PreScript:        req.PreScript,
		PostScript:       req.PostScript,
		Created:          time.Now(),
		Spec:             req.Spec,
		UUID:             helper.NewUUID().String(),
		Comment:          req.Comment,
		RequiredApproval: req.RequiredApproval,
		Hosts:            req.Hosts,
	}
	if err = t.TaskService.CreateTask(task); err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
		return
	}
	if req.RequiredApproval {
		t.cache.Set(taskPrefix+task.UUID, task, 0)
		ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: fmt.Sprintf("Task required approval, check %s/tasks/active/%s", ctx.Request.Host, taskPrefix+task.UUID)})
	} else {
		if req.Spec != "" {
			t.scheduling <- task
		} else {
			t.incoming <- task
		}
		ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: fmt.Sprintf("Task will execute very soon, check %s/tasks/events/%s for detail later", ctx.Request.Host, eventPrefix+task.UUID)})
	}
}

// field `module` must not be empty.
// field `data` will unmarshal to a predefined module
type putTaskRequest struct {
	Name             string          `json:"name"`
	PreScript        string          `json:"pre_script"`
	Module           string          `json:"module"`
	Data             json.RawMessage `json:"data"`
	PostScript       string          `json:"post_script"`
	Spec             string          `json:"spec"`
	Comment          string          `json:"comment"`
	RequiredApproval bool            `json:"required_approval"`
	Hosts            []string        `json:"hosts"`
}

// url: /tasks  method: GET
func (t *TaskHandler) getTasks(ctx *gin.Context) {
	if tasks, err := t.TaskService.Tasks(false, false); err == pub.ErrTaskSetEmpty {
		Error(ctx, err, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		ctx.IndentedJSON(http.StatusOK, tasks)
	}
}

// url: /tasks/detail/:id  method: GET
func (t *TaskHandler) getTaskByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if task, err := t.TaskService.Task(id); err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrTaskNotFound, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		ctx.IndentedJSON(http.StatusOK, task)
	}
}

// url: /tasks/detail/:id  method: POST
// use for update task's `Spec` or `Suspended`
func (t *TaskHandler) modifyTaskByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	var req postTaskRequest
	if err = ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if req.ID == 0 || req.ID != id {
		Error(ctx, pub.Error("ID is a hidden filed, equal to ID in request uri"), http.StatusBadRequest, nil)
		return
	}
	if req.Spec != "" && len(strings.Split(req.Spec, " ")) != 6 {
		Error(ctx, pub.Error("Scheduled filed is not a valid crond format"), http.StatusBadRequest, nil)
		return
	}
	task, err := t.TaskService.Task(id)
	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrTaskNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
		return
	}
	if req.Spec == task.Spec && req.Suspended == task.Suspended {
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "No fields updated"})
		return
	}
	if req.Spec != "" {
		task.Spec = req.Spec
	}
	task.Suspended = req.Suspended
	if err = t.TaskService.UpdateTask(task.ID, task); err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		t.scheduling <- task
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Update task success"})
	}
}

// update task
type postTaskRequest struct {
	ID        uint64 `json:"id"`
	Spec      string `json:"spec"`
	Suspended bool   `json:"suspended"`
}

// url: /tasks/events/:id  method: GET
func (t *TaskHandler) getTaskEventByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if evt := t.cache.Get(id); evt == nil {
		Error(ctx, pub.ErrTaskNotFound, http.StatusNotFound, nil)
	} else {
		ctx.IndentedJSON(http.StatusOK, evt)
	}
}

// url: /tasks/active/:id  method: GET
func (t *TaskHandler) activeTaskByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if evt := t.cache.Get(id); evt == nil {
		Error(ctx, pub.ErrTaskNotFound, http.StatusNotFound, nil)
	} else {
		task := evt.(*pub.Task)
		if task.Spec != "" {
			t.scheduling <- task
		} else {
			t.incoming <- task
		}
		t.cache.Delete(id)
		id := strings.Replace(id, taskPrefix, eventPrefix, -1)
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: fmt.Sprintf("Task will execute very soon, check %s/tasks/events/%s for detail later", ctx.Request.Host, id)})
	}
}

func (t *TaskHandler) initCrons() {
	crons, err := t.TaskService.Crons()
	if err != nil {
		Infof(t.Logger, "Error when getting crons: %s", err)
		return
	}
	if len(crons) > 0 {
		for _, c := range crons {
			t.cronPool <- &c
		}
	}
}

func (t *TaskHandler) runCron() {
	for {
		select {
		case item := <-t.cronPool:
			if ci := t.cache.Get(cronPrefix + item.Name); ci != nil {
				c := ci.(*pub.Cron)
				c.Suspended = true
				t.cache.Delete(cronPrefix + c.Name)
				if item.Suspended {
					break
				}
			}
			t.cache.Set(cronPrefix+item.Name, item, 0)
			t.cron.AddFunc(item.Spec, func() {
				if item.Suspended {
					return
				}
				item.Run()
				if item.Err != nil {
					Errorf(t.Logger, "Cron %s, error: %s", item.Name, item.Err.Error())
				} else {
					Infof(t.Logger, "Cron %s, output: %s", item.Name, item.Output)
				}
				t.TaskService.UpdateCron(item.ID, item)
			})
		}
	}
}

// url: /crons  method: PUT
func (t *TaskHandler) createCronJob(ctx *gin.Context) {
	var req pub.Cron
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, ErrInvalidJSON, http.StatusBadRequest, nil)
		return
	}
	cron, err := t.TaskService.CronByName(req.Name)
	if err != nil && err != pub.ErrCronNotFound {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
		return
	}
	if cron != nil {
		Error(ctx, pub.Error("Cron already exists"), http.StatusConflict, nil)
		return
	}
	req.Created = time.Now()
	if err = t.TaskService.CreateCron(&req); err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		t.cronPool <- &req
		ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: "Put cron success"})
	}
}

// url: /crons  method: GET
func (t *TaskHandler) getCronJobs(ctx *gin.Context) {
	crons, err := t.TaskService.Crons()
	if err == pub.ErrCronSetEmpty {
		Error(ctx, err, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		ctx.IndentedJSON(http.StatusOK, crons)
	}
}

// url: /crons/detail/:id  method: GET
func (t *TaskHandler) getCronJobByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	cron, err := t.TaskService.Cron(id)
	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrCronNotFound, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		ctx.IndentedJSON(http.StatusOK, cron)
	}
}

// url: /crons/detail/:id  method: POST
func (t *TaskHandler) modifyCronJobByID(ctx *gin.Context) {
	cronID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	var req postCronRequest
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if req.ID == 0 || req.ID != cronID {
		Error(ctx, pub.Error("ID is a hidden filed, equal to ID in request uri"), http.StatusBadRequest, nil)
		return
	}
	if req.Spec != "" && len(strings.Split(req.Spec, " ")) != 6 {
		Error(ctx, pub.Error("Spec filed is not a valid crond format"), http.StatusBadRequest, nil)
		return
	}
	cron, err := t.TaskService.Cron(cronID)
	if err == pub.ErrCronNotFound {
		Error(ctx, pub.ErrCronNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
		return
	}
	if req.Spec == cron.Spec && req.Suspended == cron.Suspended {
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "No fields updated"})
		return
	}
	if req.Spec != "" {
		cron.Spec = req.Spec
	}
	cron.Suspended = req.Suspended
	if err = t.TaskService.UpdateCron(cron.ID, cron); err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		t.cronPool <- cron
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Update cron success"})
	}
}

type postCronRequest struct {
	ID        uint64 `json:"id"`
	Spec      string `json:"spec"`
	Suspended bool   `json:"suspended"`
}

// url: /crons/detail/:id  method: DELETE
func (t *TaskHandler) deleteCronJobByID(ctx *gin.Context) {
	cronID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	cron, err := t.TaskService.Cron(cronID)
	if err == pub.ErrCronNotFound {
		Error(ctx, pub.ErrCronNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
		return
	}
	if err = t.TaskService.DeleteCron(cron.ID); err != nil {
		Error(ctx, err, http.StatusInternalServerError, t.Logger)
	} else {
		cron.Suspended = true
		t.cronPool <- cron
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Delete cron success"})
	}
}
