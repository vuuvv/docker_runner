package services

import (
	"bytes"
	"fmt"
	"github.com/vuuvv/docker-runner/forms"
	"github.com/vuuvv/docker-runner/services/tasks"
	"github.com/vuuvv/docker-runner/utils"
	"github.com/vuuvv/errors"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
)

type dockerService struct {
}

var DockerService = &dockerService{}

func (this *dockerService) Build(form *forms.CI) (task *tasks.Task) {
	typ := "build"
	if form.GitRevision == "" {
		form.GitRevision = "master"
	}
	if form.ImageTag == "" {
		loc := time.FixedZone("UTC+8", 8*60*60)
		form.ImageTag = time.Now().In(loc).Format("2006-01-02_15-04-05")
	}

	task = tasks.NewTask(
		tasks.WithType(typ),
		tasks.WithInput(form.GitRepository()),
		tasks.WithOutput(form.ImageRepository()),
		tasks.WithData(form),
	)

	ip, err := utils.ExternalIP()
	if err != nil {
		return task.Complete(err)
	}

	err = utils.MountSecret("/workspace/.docker/config.json", form.DockerSecret)
	if err != nil {
		return task.Complete(err)
	}
	err = utils.MountSecret("/workspace/.kube/config", form.KubeConfig)
	if err != nil {
		return task.Complete(err)
	}
	err = utils.MountSecret("/workspace/.ssh/id_rsa", form.GitSecret)
	if err != nil {
		return task.Complete(err)
	}

	cmd := exec.Command("cp", "-rf", "/app/scripts", "/workspace")

	if err = cmd.Run(); err != nil {
		return task.Complete(err)
	}

	var output bytes.Buffer

	task.Run()
	task.SetStep("Start docker container")

	cmd = exec.Command(
		"docker",
		"run",
		"-d",
		"--name", task.Id,
		"--network", fmt.Sprintf("container:%s", os.Getenv("HOSTNAME")),
		"-e", fmt.Sprintf("APP_ID=%s", task.Id),
		"-e", fmt.Sprintf("RUNNER_IP=%s", ip),
		"-e", fmt.Sprintf("GIT_URL=%s", form.GitUrl),
		"-e", fmt.Sprintf("GIT_REVISION=%s", form.GitRevision),
		"-e", fmt.Sprintf("IMAGE_URL=%s", form.ImageUrl),
		"-e", fmt.Sprintf("IMAGE_TAG=%s", form.ImageTag),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", "/workspace:/workspace",
		"-v", "/workspace/.docker/config.json:/root/.docker/config.json",
		"-v", "/workspace/.kube/config:/root/.kube/config",
		"-v", "/workspace/.ssh/id_rsa:/root/.ssh/id_rsa",
		"registry.aliyuncs.com/vuuvv/docker:20.10.8-git", "sh", "/workspace/scripts/git-clone.sh",
	)

	cmd.Stdout = &output
	cmd.Stderr = &output

	err = cmd.Run()
	if err != nil {
		return task.Complete(err)
	}

	return task
}

func (this *dockerService) GetLogsByTaskId(taskId string) (task *tasks.Task, err error) {
	task, ok := tasks.GetTask(taskId)
	if !ok {
		return nil, errors.Errorf("Not found task: %s", taskId)
	}

	return task, this.GetLogs(task)
}

func (this *dockerService) GetLogs(task *tasks.Task) (err error) {
	output, err := utils.RunCommand("docker", "logs", task.Id)
	task.Log = output
	return errors.WithStack(err)
}

func (this *dockerService) RemoveContainer(task *tasks.Task) (err error) {
	output, err := utils.RunCommand("docker", "rm", task.Id)
	if err != nil {
		task.Log += fmt.Sprintf("\n%s", output)
	} else {
		task.Log += "Remove container success"
	}

	output, err = utils.RunCommand("docker", "image", "prune", "-f")
	if err != nil {
		task.Log += fmt.Sprintf("\n%s", output)
	} else {
		task.Log += "Remove dangle image success"
	}

	return errors.WithStack(err)
}

func (this *dockerService) Complete(taskId string) (task *tasks.Task, err error) {
	task, ok := tasks.GetTask(taskId)
	if !ok {
		return nil, errors.Errorf("Not found task: %s", taskId)
	}

	// 获取日志
	err = this.GetLogs(task)
	if err != nil {
		zap.L().Error("Complete task, get log error", zap.Error(err))
	}
	// 移除container
	err = this.RemoveContainer(task)
	if err != nil {
		zap.L().Error("Complete task, remove container error", zap.Error(err))
	}
	// 删除临时目录
	err = os.RemoveAll("/workspace/" + task.Id)
	if err != nil {
		zap.L().Error("Complete task, remove tmp directory error", zap.Error(err))
	}

	task.Complete(nil)
	return task, nil
}
