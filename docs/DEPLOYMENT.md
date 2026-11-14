# OneStarLab 部署指南

目标方案：**GitHub Actions 构建镜像 → 推送阿里云容器镜像服务（ACR）→ Actions 通过 SSH 在 VPS 上执行 `docker compose pull && docker compose up -d`**。

仓库只负责构建与推送；VPS 上的 compose、`.env`、数据目录、反向代理都由你自行维护，仓库里不保存这些配置。

> 以下账号、密钥、路径均为占位符，需要你在对应平台填写。

## 0. 方案总览

```text
git tag v1.0.0
      │
      ▼
GitHub Actions (release.yml)
  1. verify   npm build + go build/vet
  2. publish  docker buildx build --push
      │        <ACR_REGISTRY>/<ACR_NAMESPACE>/onestarlab:v1.0.0
      │        <ACR_REGISTRY>/<ACR_NAMESPACE>/onestarlab:latest
      ▼
阿里云 ACR（私有仓库）
      │  docker compose pull
      ▼
VPS 上的 <APP_DIR>
  你的 compose 文件 + .env
  GitHub Actions 只执行：pull → up -d → ps
      │
      ▼
反向代理（NPM 等）→ https://<域名>
```

设计边界：

- Actions 不保存 VPS 上的 compose 内容，只执行固定的 compose 命令。
- 镜像由 tag 决定；回滚就是把 `IMAGE_TAG` 指回旧版本再部署一次。
- 自动健康检查和自动回滚不在 workflow 里，需要的话放在你自己的 compose healthcheck 或 VPS 脚本里。

## 1. 阿里云 ACR

### 1.1 创建实例与仓库

1. 登录阿里云控制台，进入「容器镜像服务 ACR」。
2. 选择个人版或企业版实例。
3. 创建命名空间，例如 `<ACR_NAMESPACE>`。
4. 在命名空间下创建**私有**仓库，名称必须是 `onestarlab`（与 workflow 中的仓库名一致）。
5. 记下实例的**公网地址**，形如：
   - 个人版：`registry.cn-hangzhou.aliyuncs.com`
   - 企业版：以控制台显示为准，通常为 `<实例名>-registry.<region>.cr.aliyuncs.com`

该地址就是 `ACR_REGISTRY`，不要带 `https://`。

### 1.2 创建 RAM 用户并授权

1. 在 RAM 访问控制创建用户，勾选「OpenAPI 调用访问」。
2. 授予 ACR 推拉权限：`AliyunContainerRegistryFullAccess`，或只包含目标命名空间的自定义策略。
3. 在 ACR 的「访问凭证」页面为该用户设置**固定密码**，这就是 `ACR_PASSWORD`。
4. `ACR_USERNAME` 填 RAM 用户名，以控制台显示为准。

### 1.3 VPS 拉取权限

VPS 也需要登录 ACR。可以复用同一个 RAM 用户，或单独建一个只读账号。登录一次后凭据保存在该用户的 `~/.docker/config.json`。

如果使用企业版并绑定了 VPC：GitHub 托管 runner 只能用公网地址推送；VPS 可以在 compose 里使用 VPC 地址拉取，但需要单独对 VPC 地址执行一次 `docker login`。

## 2. VPS 侧准备

仓库不提供 compose 文件，你需要自己维护一份，放在某个目录下（下称 `<APP_DIR>`）。

### 2.1 基本要求

- 已安装 Docker 与 compose 插件。
- 执行部署的用户在 `docker` 组，或能通过 sudo 执行 docker。
- 已存在一个供反向代理使用的 external 网络，例如 `proxy`。
- `<APP_DIR>` 下有 compose 文件（`compose.yaml` 或 `docker-compose.yml`）和 `.env`。
- 数据目录可被容器内运行用户写入；镜像内应用以 `uid=100` 运行。

### 2.2 compose 需要满足的约定

- 镜像来自 ACR：`<ACR_REGISTRY>/<ACR_NAMESPACE>/onestarlab:${IMAGE_TAG}`。
- 必须使用 `${IMAGE_TAG}` 插值，workflow 会在远端命令里传入该变量。
- SQLite 数据目录挂载到容器内 `/data`。
- 容器内监听 `8080`；建议不映射宿主机端口，加入 external 网络由 NPM 反代。
- 需要的话自行配置 `healthcheck`、日志轮转、`restart: unless-stopped`。

Compose 示例结构（占位符自行替换，不要提交到仓库）：

```yaml
services:
  onestarlab:
    image: <ACR_REGISTRY>/<ACR_NAMESPACE>/onestarlab:${IMAGE_TAG:-latest}
    container_name: <容器名>
    environment:
      PORT: "8080"
      DATABASE_PATH: /data/onestar.db
      WEB_DIR: /app/web
    volumes:
      - <数据目录>:/data
    networks:
      - proxy
    restart: unless-stopped

networks:
  proxy:
    external: true
```

### 2.3 环境变量

在 `.env` 或 compose 中配置：

```dotenv
IMAGE_TAG=latest
API_TOKEN=
CORS_ORIGINS=*
TZ=Asia/Shanghai
```

说明：

- `IMAGE_TAG` 由 workflow 传入，`.env` 里的值只作为手动执行时的默认值。
- `API_TOKEN` 建议留空。前端 token 是构建时注入的，当前镜像没有注入；服务端设了非空值页面会 401。访问控制优先放在 HTTPS/反向代理或安全组。
- 不要把 `.env` 提交到仓库。

### 2.4 首次登录与手动验证

```bash
cd <APP_DIR>
docker login --username='<RAM用户名>' <ACR_REGISTRY>
IMAGE_TAG=latest docker compose pull
```

镜像还没推送过时这一步会失败，等 Actions 发布后再试。

### 2.5 防火墙

只放行 `22`、`80`、`443`。本项目不需要暴露宿主机 8080。

## 3. GitHub 配置

### 3.1 Secrets

仓库 → Settings → Secrets and variables → Actions → New repository secret：

| Secret | 说明 |
| --- | --- |
| `ACR_REGISTRY` | ACR 公网地址，不带协议 |
| `ACR_NAMESPACE` | ACR 命名空间 |
| `ACR_USERNAME` | RAM 用户名 |
| `ACR_PASSWORD` | ACR 访问凭证的固定密码 |
| `VPS_HOST` | VPS 公网 IP 或域名 |
| `VPS_PORT` | SSH 端口，留空默认 22 |
| `VPS_USER` | 部署用户（你在 VPS 上使用的用户） |
| `VPS_SSH_KEY` | 私钥全文，含 BEGIN/END 行 |
| `VPS_KNOWN_HOSTS` | `ssh-keyscan` 输出；为空则部署时自动扫描 |
| `VPS_APP_DIR` | VPS 上存放 compose 文件的目录 |

`VPS_APP_DIR` 是仓库里唯一提到的路径，但它的值只存在于 Secrets 中。

可选 Repository variable：

| Variable | 说明 |
| --- | --- |
| `TARGET_PLATFORM` | 构建平台，默认 `linux/amd64`；ARM 改成 `linux/arm64` |

### 3.2 production 环境（可选）

Settings → Environments → New environment → `production`，可添加 Required reviewers，让部署前需要人工批准。

### 3.3 触发发布

```bash
git tag -a v1.0.0 -m "release v1.0.0"
git push origin v1.0.0
```

工作流：

1. `verify`：前端构建 + 后端 `go build`/`go vet`
2. `publish`：构建并推送 `v1.0.0`、`latest`、buildcache
3. `deploy`：SSH 到 VPS，在 `VPS_APP_DIR` 执行 `docker compose pull && up -d --remove-orphans && ps`

也可以手动触发：Actions → Release → Run workflow。

- `tag` 留空时使用 `sha-<7位提交号>`。
- 首次只验证 ACR 推送时，勾选 `skip_deploy`，不会碰 VPS。

> 手动触发要求 `release.yml` 已存在于默认分支。

## 4. 回滚

workflow 不再自动回滚。回滚方式：

1. 在 VPS 的 `<APP_DIR>` 手动执行：

   ```bash
   IMAGE_TAG=v1.0.0 docker compose pull
   IMAGE_TAG=v1.0.0 docker compose up -d
   ```

2. 或在 Actions 手动触发，`tag` 填旧版本号（会覆盖 `IMAGE_TAG` 并部署）。

ACR 的保留策略可能清理旧 tag，重要版本请保留。

## 5. 反向代理与 HTTPS

容器建议只加入 external 网络，不映射宿主机端口。

### 5.1 共享网络

```bash
docker network ls | grep proxy || docker network create proxy
```

如果 NPM 已经接入某个共享网络，直接用那个名字。

### 5.2 NPM 接入

NPM 的 compose 里声明同一个 external 网络，或临时执行：

```bash
docker network connect proxy <npm容器名>
```

临时连接在 NPM 容器重建后会丢失。

### 5.3 Proxy Host

- Domain Names：你的域名
- Scheme：`http`
- Forward Hostname：compose 里的容器名或服务名
- Forward Port：`8080`
- SSL 页签申请证书并开启 Force SSL

### 5.4 验证

```bash
docker network inspect proxy --format '{{range .Containers}}{{.Name}} {{end}}'
docker exec <容器名> wget -qO- http://127.0.0.1:8080/api/v1/health
```

## 6. 备份与恢复

VPS 上安装 `sqlite3` 后，用 SQLite 在线备份：

```bash
cd <APP_DIR>
mkdir -p backups
sqlite3 data/onestar.db ".backup 'backups/onestar-$(date +%Y%m%d-%H%M%S).db'"
gzip backups/onestar-*.db
```

加入 crontab 定期执行，并把 `backups/` 同步到其他机器或对象存储。

恢复：

```bash
cd <APP_DIR>
docker compose stop
gunzip -c backups/<备份文件>.db.gz > data/onestar.db
docker compose start
```

也可以继续用页面上的「数据中心 → 导入数据」恢复 JSON 快照，不需要停容器。

## 7. 常见问题

### `docker login` 失败或推送 401

- 确认 `ACR_REGISTRY` 与控制台一致。
- 确认使用的是 ACR 固定密码，不是阿里云登录密码。
- 确认 RAM 用户有目标命名空间的推拉权限。
- 本地验证：

  ```bash
  docker login --username='<RAM用户名>' <ACR_REGISTRY>
  docker pull <ACR_REGISTRY>/<ACR_NAMESPACE>/onestarlab:latest
  ```

### deploy 步骤 SSH 失败

- `Permission denied`：私钥与 `authorized_keys` 不匹配，或用户名/端口错误。
- `Host key verification failed`：重新生成 `VPS_KNOWN_HOSTS`。
- 连接超时：安全组/防火墙未放行 SSH 端口。

### deploy 步骤报 `VPS_APP_DIR` 相关错误

- 确认 secret 已配置，且路径是绝对路径。
- 确认该目录下存在 `compose.yaml` 或 `docker-compose.yml`。
- 确认部署用户对目录有读权限、能执行 docker。

### `docker compose pull` 失败

- VPS 是否已 `docker login`。
- ACR 仓库是否私有、账号是否有拉取权限。
- `no matching manifest`：镜像架构与 VPS 不一致，设置 `TARGET_PLATFORM`。

### 容器一直 unhealthy

```bash
cd <APP_DIR>
docker compose logs --tail 100
```

重点看数据目录权限（容器内 `uid=100` 需可写）、端口占用、数据库文件。

### NPM 502

- 容器与 NPM 是否在同一网络：

  ```bash
  docker network inspect proxy --format '{{range .Containers}}{{.Name}} {{end}}'
  ```

- NPM 的 Forward Hostname 是否填容器名、端口 `8080`、协议 `http`。
- 应用本身是否正常：`docker exec <容器名> wget -qO- http://127.0.0.1:8080/api/v1/health`。

## 8. 安全检查清单

- [ ] ACR 仓库为私有，RAM 用户最小权限
- [ ] ACR 密码定期轮换，不写入仓库
- [ ] VPS 使用专用 SSH key，禁用密码登录
- [ ] `production` 环境开启 Required reviewers
- [ ] 安全组只放行 22/80/443
- [ ] NPM 与容器同网络，8080 不直接暴露
- [ ] 配置 HTTPS 后再对外使用
- [ ] 每日备份并验证可恢复
- [ ] 定期更新系统与 Docker
