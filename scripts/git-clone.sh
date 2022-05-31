#!/usr/bin/env sh
set -ex

git config --global init.defaultBranch master
git config --global advice.detachedHead false

if test -z ${APP_ID}; then
  echo 'ERROR: environment "APP_ID" not set!'
  exit 1
fi

if test -z ${GIT_URL}; then
  echo 'ERROR: environment "GIT_URL" not set!'
  exit 1
fi

if test -z ${IMAGE_URL}; then
  echo 'ERROR: environment "IMAGE_URL" not set!'
  exit 1
fi

if test -z ${IMAGE_TAG}; then
  echo 'ERROR: environment "IMAGE_TAG" not set!'
  exit 1
fi

if test -z ${GIT_REVISION}; then
  GIT_REVISION="master"
fi

if test -z ${WORKSPACE}; then
  WORKSPACE="/workspace/${APP_ID}"
fi

if test -z ${BUILD_DIRECTORY}; then
  BUILD_DIRECTORY="."
fi

if test -z ${DOCKERFILE}; then
  DOCKERFILE="Dockerfile"
fi


CHECKOUT_DIR="${WORKSPACE}/code"

cleandir() {
  # Delete any existing contents of the repo directory if it exists.
  #
  # We don't just "rm -rf ${CHECKOUT_DIR}" because ${CHECKOUT_DIR} might be "/"
  # or the root of a mounted volume.
  if [ -d "${CHECKOUT_DIR}" ] ; then
    # Delete non-hidden files and directories
    rm -rf "${CHECKOUT_DIR:?}"/*
    # Delete files and directories starting with . but excluding ..
    rm -rf "${CHECKOUT_DIR}"/.[!.]*
    # Delete files and directories starting with .. plus any other character
    rm -rf "${CHECKOUT_DIR}"/..?*
  fi
}

step() {
  STEP_DATA=$(printf '{"taskId": "%s", "step":"%s"}' "${APP_ID}" "$1")
  RESP=$(curl --no-progress-meter -X POST "http://${RUNNER_IP}:3000/docker/step" -H 'Content-Type: application/json' -d "${STEP_DATA}")
  echo "${RESP}"
}

step "Create checkout directory"
cleandir

mkdir -p ${CHECKOUT_DIR}

#echo "TASK_ID: ${APP_ID}"
#echo "CHECKOUT_DIR: ${CHECKOUT_DIR}"
#echo "GIT_URL: ${GIT_URL}"
#echo "GIT_REVISION: ${GIT_REVISION}"
#echo "IMAGE_URL: ${IMAGE_URL}"
#echo "IMAGE_TAG: ${IMAGE_TAG}"

#echo "Change directory: ${CHECKOUT_DIR}"
cd ${CHECKOUT_DIR}

#echo "Pull code: ${GIT_URL} ${GIT_REVISION}"
step "Pull code"
git init .

git remote add ${APP_ID} ${GIT_URL}

git fetch --recurse-submodules=yes --depth=1 ${APP_ID} --update-head-ok --force ${GIT_REVISION}

COMMIT_ID=$(git show -q --pretty=format:%H FETCH_HEAD)

git checkout -f ${COMMIT_ID}

step "Build image"
cd ${BUILD_DIRECTORY}
IMAGE="${IMAGE_URL}:${IMAGE_TAG}"
#echo "Build image: ${IMAGE}"
docker build -f ${DOCKERFILE} -t ${IMAGE} .

#echo "Push image: ${IMAGE}"
docker push ${IMAGE}

if [ ${IMAGE_TAG} != "latest" ]; then
  LATEST_IMAGE="${IMAGE_URL}:latest"
  docker tag ${IMAGE} ${LATEST_IMAGE}
  docker push ${LATEST_IMAGE}
fi

#echo "Clean image"
docker image prune -f
docker image rm -f ${IMAGE}

step "Notify success"
#echo "Notify success"
RESP=$(curl --no-progress-meter -X POST "http://${RUNNER_IP}:3000/docker/complete?id=${APP_ID}")
echo ${RESP}

echo "Success"

