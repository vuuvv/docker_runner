package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/vuuvv/docker-runner/forms"
	"github.com/vuuvv/docker-runner/services/tasks"
	"github.com/vuuvv/docker-runner/utils"
	"github.com/vuuvv/errors"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
	"vuuvv.cn/unisoftcn/orca"
	"vuuvv.cn/unisoftcn/orca/serialize"
	orcaUtils "vuuvv.cn/unisoftcn/orca/utils"
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

	cloneForm := *form
	cloneForm.DockerSecret = ""
	cloneForm.GitSecret = ""
	task = tasks.NewTask(
		tasks.WithType(typ),
		tasks.WithInput(form.GitRepository()),
		tasks.WithOutput(form.ImageRepository()),
		tasks.WithData(cloneForm),
	)

	ip, err := utils.ExternalIP()
	if err != nil {
		return task.Complete(err)
	}

	workspace := fmt.Sprintf("/workspace/%s", task.Id)
	if err = os.MkdirAll(workspace, 0755); err != nil {
		return task.Complete(err)
	}
	secretPath := fmt.Sprintf("%s/secrets", workspace)
	if err = os.MkdirAll(secretPath, 0755); err != nil {
		return task.Complete(err)
	}

	err = utils.MountSecret(fmt.Sprintf("%s/.docker/config.json", secretPath), form.DockerSecret, 0400)
	if err != nil {
		return task.Complete(err)
	}
	err = utils.MountSecret(fmt.Sprintf("%s/.kube/config", secretPath), form.KubeConfig, 0400)
	if err != nil {
		return task.Complete(err)
	}
	err = utils.MountSecret(fmt.Sprintf("%s/.ssh/id_rsa", secretPath), form.GitSecret, 0400)
	if err != nil {
		return task.Complete(err)
	}

	cmd := exec.Command("cp", "-rf", "/app/scripts", workspace)

	if err = cmd.Run(); err != nil {
		return task.Complete(err)
	}

	var output bytes.Buffer

	task.Run()
	task.SetStep("Start docker container")

	zap.L().Info("start docker container")
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
		"-e", fmt.Sprintf("BUILD_DIRECTORY=%s", form.BuildDirectory),
		"-e", fmt.Sprintf("DOCKERFILE=%s", form.Dockerfile),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", fmt.Sprintf("%s:/workspace", workspace),
		"-v", fmt.Sprintf("%s/.docker/config.json:/root/.docker/config.json", secretPath),
		"-v", fmt.Sprintf("%s/.kube/config:/root/.kube/config", secretPath),
		"-v", fmt.Sprintf("%s/.ssh/id_rsa:/root/.ssh/id_rsa", secretPath),
		"-v", fmt.Sprintf("%s/scripts/ssh_config:/root/.ssh/config", workspace),
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

// StopContainer 停止容器并清除现场
func (this *dockerService) StopContainer(task *tasks.Task) (err error) {
	output, err := utils.RunCommand("docker", "stop", task.Id)
	if err != nil {
		task.CleanLog += fmt.Sprintf("\n%s", output)
	} else {
		task.CleanLog += "Stop container success"
	}

	return this.RemoveContainer(task)
}

// RemoveContainer 正常的清除现场工作
func (this *dockerService) RemoveContainer(task *tasks.Task) (err error) {
	output, err := utils.RunCommand("docker", "rm", task.Id)
	if err != nil {
		task.CleanLog += fmt.Sprintf("\n%s", output)
	} else {
		task.CleanLog += "Remove container success"
	}

	output, err = utils.RunCommand("docker", "image", "prune", "-f")
	if err != nil {
		task.CleanLog += fmt.Sprintf("\n%s", output)
	} else {
		task.CleanLog += "\nRemove dangle image success"
	}

	return errors.WithStack(err)
}

// WatchTask 监控所有任务状态, 如果是完成状态，清理任务现场并向消息队列发送任务成功消息
func (this *dockerService) WatchTask() {
	c := cron.New()
	_, err := c.AddFunc("@every 3s", func() {
		defer orcaUtils.NormalRecover("Watch task")
		taskList := tasks.GetTaskList()
		this.FetchContainerInfo(taskList...)
		for _, task := range taskList {
			if task.IsContainerExited() {
				this.CleanUp(task)
				zap.L().Info("Notify completed task", zap.String("taskId", task.Id))
				err := orca.Redis().Publish(context.Background(), "/docker-runner/task/completed", task.Id).Err()
				if err != nil {
					task.RelateError = err
					zap.L().Error("Publish completed message error", zap.Error(err))
					continue
				}
			}
		}
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}

func (this *dockerService) FetchContainerInfo(taskList ...*tasks.Task) {
	for _, task := range taskList {
		if task.IsClean {
			continue
		}
		output, err := utils.RunCommand("docker", "inspect", task.Id)
		if err != nil {
			task.ContainerInfo = nil
			task.RelateError = err
			zap.L().Error(err.Error(), zap.Error(err), zap.String("cmd", "docker inspect "+task.Id))
			continue
		}

		infoList, err := serialize.JsonParsePrimitive[[]*tasks.ContainerInfo](output)
		if err != nil {
			task.ContainerInfo = nil
			task.RelateError = err
			zap.L().Error(err.Error(), zap.Error(err))
			continue
		}
		if len(infoList) == 0 {
			task.ContainerInfo = nil
			task.RelateError = errors.Errorf("Can not get docker container info: %s", task.Id)
			zap.L().Error(err.Error(), zap.Error(task.RelateError))
			continue
		}
		task.ContainerInfo = infoList[0]
	}
}

func (this *dockerService) CleanUp(task *tasks.Task) {
	if task.IsClean {
		return
	}
	// 获取日志
	err := this.GetLogs(task)
	if err != nil {
		task.RelateError = err
		zap.L().Error("Complete task, get log error", zap.Error(err))
	}
	// 移除container
	err = this.RemoveContainer(task)
	if err != nil {
		task.RelateError = err
		zap.L().Error("Complete task, remove container error", zap.Error(err))
	}
	// 删除临时目录
	err = os.RemoveAll("/workspace/" + task.Id)
	if err != nil {
		task.CleanLog += err.Error()
		task.RelateError = err
		zap.L().Error("Complete task, remove tmp directory error", zap.Error(err))
	}
	task.IsClean = true
}

func (this *dockerService) Complete(taskId string) (task *tasks.Task, err error) {
	task, ok := tasks.GetTask(taskId)
	if !ok {
		return nil, errors.Errorf("Not found task: %s", taskId)
	}
	task.Complete(nil)
	return task, nil
}
