package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/vuuvv/docker-runner/forms"
	"github.com/vuuvv/docker-runner/services"
	"github.com/vuuvv/docker-runner/services/tasks"
	"github.com/vuuvv/docker-runner/utils"
	"github.com/vuuvv/errors"
	"vuuvv.cn/unisoftcn/orca/base"
	"vuuvv.cn/unisoftcn/orca/server"
)

type dockerController struct {
	server.BaseController
}

var DockerController = &dockerController{}

func (this *dockerController) Name() string {
	return "docker"
}

func (this *dockerController) Path() string {
	return "docker"
}

func (this *dockerController) Middlewares() []gin.HandlerFunc {
	return nil
}

func (this *dockerController) Mount(router *gin.RouterGroup) {
	this.Post("exec", this.exec)
	this.Post("ci", this.ci)
	this.Post("step", this.step)
	this.Post("complete", this.complete)
	this.Post("stop", this.stop)
	this.Get("task", this.getTask)
	this.Delete("task", this.deleteTask)
	this.Get("tasks", this.tasks)
	this.Get("task_map", this.taskMap)
	this.Get("log", this.log)
	this.Get("ip", this.ip)
}

func (this *dockerController) exec(ctx *gin.Context) {
}

func (this *dockerController) ci(ctx *gin.Context) {
	form := &forms.CI{}
	err := this.ValidForm(form)
	if err != nil {
		this.SendError(err)
		return
	}
	this.Send(services.DockerService.Build(form))
}

func (this *dockerController) ip(ctx *gin.Context) {
	this.SendW(utils.ExternalIP())
}

func (this *dockerController) tasks(ctx *gin.Context) {
	this.Send(tasks.GetTaskList())
}

func (this *dockerController) taskMap(ctx *gin.Context) {
	this.Send(tasks.GetTaskMap())
}

func (this *dockerController) log(ctx *gin.Context) {
	id, ok := ctx.GetQuery("id")
	if !ok {
		this.SendError(errors.New("请传入task id"))
		return
	}
	this.SendW(services.DockerService.GetLogsByTaskId(id))
}

func (this *dockerController) getTask(ctx *gin.Context) {
	id, ok := ctx.GetQuery("id")
	if !ok {
		this.SendError(errors.New("请传入task id"))
		return
	}
	task, ok := tasks.GetTask(id)
	if !ok {
		this.SendError(errors.Errorf("任务不存在：%s", id))
		return
	}
	this.Send(task)
}

func (this *dockerController) deleteTask(ctx *gin.Context) {
	id, ok := ctx.GetQuery("id")
	if !ok {
		this.SendError(errors.New("请传入task id"))
		return
	}
	this.SendW(tasks.DeleteTask(id))
}

func (this *dockerController) complete(ctx *gin.Context) {
	id, ok := ctx.GetQuery("id")
	if !ok {
		this.SendError(errors.New("请传入task id"))
		return
	}
	task, ok := tasks.GetTask(id)
	if !ok {
		this.SendError(errors.New("任务不存在"))
		return
	}
	_, err := services.DockerService.Complete(id)
	if err != nil {
		this.SendError(err)
		return
	}
	this.Send(task)
}

func (this *dockerController) stop(ctx *gin.Context) {
	id, ok := ctx.GetQuery("id")
	if !ok {
		this.SendError(errors.New("请传入task id"))
		return
	}
	task, ok := tasks.GetTask(id)
	if !ok {
		this.SendError(errors.New("任务不存在"))
		return
	}
	err := services.DockerService.StopContainer(task)
	if err != nil {
		this.SendError(err)
		return
	}
	this.Send(task)
}

func (this *dockerController) step(ctx *gin.Context) {
	form := &forms.Step{}
	err := this.ValidForm(form)
	if err != nil {
		this.SendError(err)
		return
	}

	task, ok := tasks.GetTask(form.TaskId)
	if !ok {
		this.SendError(base.ErrorBadRequest("任务不存在: %s", form.TaskId))
		return
	}
	task.SetStep(form.Step)

	this.Send(form.Step)
}
