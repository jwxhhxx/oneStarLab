#!/usr/bin/env bash
# 在 VPS 的 /opt/onestarlab 下执行：./deploy.sh [镜像标签]
# 不带参数时使用 .env 里的 IMAGE_TAG；GitHub Actions 会把新标签作为第一个参数传入。
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/onestarlab}"
COMPOSE_FILE="${COMPOSE_FILE:-$APP_DIR/docker-compose.yml}"
ENV_FILE="${ENV_FILE:-$APP_DIR/.env}"
CONTAINER_NAME="${CONTAINER_NAME:-onestarlab}"
HEALTH_RETRIES="${HEALTH_RETRIES:-30}"

log() { printf '[deploy] %s\n' "$*"; }
die() { printf '[deploy] ERROR: %s\n' "$*" >&2; exit 1; }

cd "$APP_DIR" 2>/dev/null || die "目录不存在：$APP_DIR"
[ -f "$COMPOSE_FILE" ] || die "缺少编排文件：$COMPOSE_FILE"
[ -f "$ENV_FILE" ] || die "缺少环境文件：$ENV_FILE（从 .env.example 复制并填写）"

TAG="${1:-${IMAGE_TAG:-latest}}"
case "$TAG" in
  ''|*[!A-Za-z0-9._-]*) die "非法的镜像标签：$TAG" ;;
esac

PREV_TAG="$(sed -n 's/^IMAGE_TAG=//p' "$ENV_FILE" | tail -n1)"
export IMAGE_TAG="$TAG"

mkdir -p "$APP_DIR/data"
# 容器以 uid=100(app)、gid=101(app) 运行；宿主机目录需要可写。
if [ "$(id -u)" -eq 0 ]; then
  chown -R 100:101 "$APP_DIR/data"
fi

log "目标版本：$TAG（上一版本：${PREV_TAG:-无}）"
docker compose -f "$COMPOSE_FILE" config -q
docker compose -f "$COMPOSE_FILE" pull

if ! docker compose -f "$COMPOSE_FILE" up -d --remove-orphans; then
  log "启动失败，尝试回滚"
  if [ -n "$PREV_TAG" ] && [ "$PREV_TAG" != "$TAG" ]; then
    IMAGE_TAG="$PREV_TAG" docker compose -f "$COMPOSE_FILE" up -d --remove-orphans || true
    log "已尝试回滚到 $PREV_TAG"
  fi
  docker compose -f "$COMPOSE_FILE" logs --tail 80 "$CONTAINER_NAME" || true
  exit 1
fi

log "等待健康检查"
healthy=0
for _ in $(seq 1 "$HEALTH_RETRIES"); do
  cid="$(docker compose -f "$COMPOSE_FILE" ps -q "$CONTAINER_NAME" 2>/dev/null || true)"
  if [ -n "$cid" ]; then
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$cid" 2>/dev/null || echo unknown)"
    if [ "$status" = "healthy" ]; then
      healthy=1
      break
    fi
  fi
  sleep 2
done

if [ "$healthy" -ne 1 ]; then
  log "健康检查未通过，尝试回滚"
  docker compose -f "$COMPOSE_FILE" logs --tail 80 "$CONTAINER_NAME" || true
  if [ -n "$PREV_TAG" ] && [ "$PREV_TAG" != "$TAG" ]; then
    IMAGE_TAG="$PREV_TAG" docker compose -f "$COMPOSE_FILE" up -d --remove-orphans || true
    log "已尝试回滚到 $PREV_TAG"
  else
    log "没有可回滚的上一版本，请人工检查"
  fi
  exit 1
fi

if grep -q '^IMAGE_TAG=' "$ENV_FILE"; then
  sed -i "s|^IMAGE_TAG=.*|IMAGE_TAG=${TAG}|" "$ENV_FILE"
else
  printf 'IMAGE_TAG=%s\n' "$TAG" >> "$ENV_FILE"
fi

docker image prune -f >/dev/null 2>&1 || true
log "部署成功：$TAG"
