package tasks

import (
	"github.com/vuuvv/docker-runner/utils"
	"sync"
	"time"
)

const (
	TaskStatusInit      = "init"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
)

var taskMap sync.Map

type Task struct {
	Id        string      `json:"id"`
	Type      string      `json:"type"`
	Input     string      `json:"input"`
	Output    string      `json:"output"`
	Status    string      `json:"status"`
	Log       string      `json:"log"`
	Error     error       `json:"error"`
	Step      string      `json:"step"`
	Result    string      `json:"result"`
	CreatedAt *time.Time  `json:"createdAt"`
	EndAt     *time.Time  `json:"endAt"`
	Data      interface{} `json:"data"`
}

func (this *Task) Complete(err error) *Task {
	this.Error = err
	this.Status = TaskStatusCompleted
	this.Step = TaskStatusCompleted
	if this.Error != nil {
		this.Result = err.Error()
	} else {
		this.Result = TaskStatusCompleted
	}
	return this
}

func (this *Task) Run() *Task {
	this.Status = TaskStatusRunning
	return this
}

func (this *Task) SetStep(step string) *Task {
	this.Step = step
	return this
}

type TaskOption func(*Task)

func WithType(typ string) TaskOption {
	return func(task *Task) {
		task.Type = typ
	}
}

func WithInput(input string) TaskOption {
	return func(task *Task) {
		task.Input = input
	}
}

func WithOutput(output string) TaskOption {
	return func(task *Task) {
		task.Output = output
	}
}

func WithData(data interface{}) TaskOption {
	return func(task *Task) {
		task.Data = data
	}
}

func NewTask(opts ...TaskOption) *Task {
	now := time.Now()
	task := &Task{
		Status:    TaskStatusInit,
		Step:      TaskStatusInit,
		CreatedAt: &now,
	}

	task.Id = utils.RandString(8)
	for {
		if _, exists := taskMap.LoadOrStore(task.Id, task); !exists {
			break
		}
		task.Id = utils.RandString(8)
	}

	for _, opt := range opts {
		opt(task)
	}
	return task
}

func GetTask(id string) (task *Task, ok bool) {
	obj, ok := taskMap.Load(id)
	if ok {
		return obj.(*Task), true
	}
	return nil, false
}

func GetTaskMap() (tasks map[string]*Task) {
	tasks = map[string]*Task{}
	taskMap.Range(func(key, value interface{}) bool {
		tasks[key.(string)] = value.(*Task)
		return true
	})
	return tasks
}

func GetTaskList() (tasks []*Task) {
	taskMap.Range(func(key, value interface{}) bool {
		tasks = append(tasks, value.(*Task))
		return true
	})
	return tasks
}
