package forms

type Step struct {
	TaskId string `json:"taskId" valid:"required~请传入taskId"`
	Step   string `json:"step" valid:"required~请传入step"`
}
