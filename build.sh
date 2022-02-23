VERSION=registry.aliyuncs.com/vuuvv/docker-runner:0.0.1
docker build -t ${VERSION} .
docker push ${VERSION}
