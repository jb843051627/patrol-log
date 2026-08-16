#!/usr/bin/env bash
set -euo pipefail

NAME="${1:?用法: build_benzhi_docker.sh <镜像名> <平台>}"
PLATFORM="${2:-linux/amd64}"
DOCKERFILE="benzhi.Dockerfile"

IMAGE="benzhi/${NAME}:latest"

docker build \
  --platform "${PLATFORM}" \
  -f "${DOCKERFILE}" \
  -t "${IMAGE}" \
  .

echo "已构建 ${IMAGE} (${PLATFORM})"
echo "运行: docker run --rm -it ${IMAGE} sh"
echo "双架构验收: 分别以 linux/amd64 与 linux/arm64 各构建并运行一次"
