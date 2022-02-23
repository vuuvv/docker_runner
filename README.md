### deploy image

generate-name

env:
  WORKSPACE
  GIT_URL
  GIT_REVISION

echo 'clean dir'
rm -rf /workspace/generate-name
mkdir -p /workspace/generate-name

docker run \
    -e APP_ID=xxx \
    -e GIT_URL=xxx \
    -e GIT_REVISION=xxx \
    -e WORKSPACE=xxx \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /workspace:/workspace \
    -v /scripts:/scripts \
    -w /workspace \
    docker:20.10.8-git \
    sh /scripts/git-clone.sh

### secrets

#### ssh private key

directory: $HOME/.ssh/id_rsa

#### docker auth config

directory: $HOME/.docker/config

#### kubeconfig

directory: $HOME/.kube/config