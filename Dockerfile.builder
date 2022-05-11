FROM golang:1.18.1

# 环境变量设置，使用私有仓库
RUN go env -w GOPRIVATE=vuuvv.cn
RUN git config --global url."git@vuuvv.cn:".insteadof "https://vuuvv.cn/"
RUN go env -w GOINSECURE=vuuvv.cn

# 启用go module
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go mod download -x

RUN GOOS=linux GOARCH=amd64 go build -o main .

RUN cd /

RUN rm -rf /app
