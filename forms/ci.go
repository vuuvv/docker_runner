package forms

import "fmt"

type CI struct {
	GitUrl       string `json:"gitUrl" valid:"required~请传入gitUrl"`
	GitRevision  string `json:"gitRevision"`
	GitSecret    string `json:"gitSecret"`
	ImageUrl     string `json:"imageUrl" valid:"required~请传入imageUrl"`
	ImageTag     string `json:"imageTag"`
	DockerSecret string `json:"dockerSecret"`
	KubeConfig   string `json:"kubeConfig"`
}

func (this *CI) GitRepository() string {
	return fmt.Sprintf("%s:%s", this.GitUrl, this.GitRevision)
}

func (this *CI) ImageRepository() string {
	return fmt.Sprintf("%s:%s", this.ImageUrl, this.ImageTag)
}
