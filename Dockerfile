FROM registry.aliyuncs.com/vuuvv/orca:0.0.9 as builder

# 启用go module
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

#RUN go mod download -x

RUN go mod tidy

RUN GOOS=linux GOARCH=amd64 go build -o main .

RUN mkdir dist && cp main dist && cp -r resources dist && cp -r scripts dist

#FROM registry.aliyuncs.com/vuuvv/base-debian11:latest
FROM docker:20.10.8-git

WORKDIR /app

COPY --from=builder /app/dist /app

ENV GIN_MODE=release

ENTRYPOINT ["./main"]