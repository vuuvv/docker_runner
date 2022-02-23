FROM golang:1.17-alpine

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
