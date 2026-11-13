# OneStarLab 部署指南

本文描述目标生产方案：**GitHub Actions 构建镜像 → 推送到阿里云容器镜像服务（ACR）→ Actions 通过 SSH 触发 VPS 上的 `docker compose` 拉取并重启**。

当前代码已经包含：

- 发布工作流：[`.github/workflows/release.yml`](../.github/workflows/release.yml)
- 生产编排：[`deploy/docker-compose.yml`](../deploy/docker-compose.yml)
- 部署脚本：[`deploy/deploy.sh`](../deploy/deploy.sh)（带健康检查与自动回滚）
- 备份脚本：[`deploy/backup.sh`](../deploy/backup.sh)
- 环境变量模板：[`deploy/.env.example`](../deploy/.env.example)

> 以下阿里云 / GitHub / VPS 的账号与密钥配置需要你手工完成；本仓库只准备好了代码和脚本。

## 0. 方案总览

```text
git tag v1.0.0
      │
      ▼
GitHub Actions (release.yml)
  1. verify   npm build + go build/vet
  2. publish  docker buildx build --push
      │        registry.cn-hangzhou.aliyuncs.com/<namespace>/onestarlab:v1.0.0
      │        registry.cn-hangzhou.aliyuncs.com/<namespace>/onestarlab:latest
      ▼
阿里云 ACR（私有仓库）
      │  docker compose pull
      ▼
VPS /opt/onestarlab
  deploy.sh：pull → up -d → 等待 healthy
             失败则回滚到 IMAGE_TAG 记录的上一版本
      │
      ▼
http://<VPS>:8080  （建议再套 Caddy/Nginx 上 HTTPS）
```

设计要点：

- 镜像在 GitHub 构建，VPS 不需要 Node、Go 或源码，只装 Docker。
- 发布以 tag 为准，镜像同时打 `v1.0.0`、`latest` 和构建缓存标签；回滚就是重新部署旧 tag。
- 部署前先 `pull`，启动后轮询容器健康状态；失败自动切回 `.env` 里记录的上一版本。
- SQLite 通过宿主机 `./data` 目录持久化，容器重建不丢数据。

## 1. 阿里云 ACR 准备

### 1.1 创建实例与仓库

1. 登录阿里云控制台，进入「容器镜像服务 ACR」。
2. 选择个人版或企业版实例：
   - 个人版：免费、配置简单，适合单机自用。
   - 企业版：支持 VPC 访问、镜像安全扫描、保留策略，适合长期使用。
3. 在实例中创建命名空间，例如 `onestarlab`。
4. 在命名空间下创建仓库：
   - 仓库名称：`onestarlab`
   - 类型：**私有**
5. 在仓库的「基本信息」里记下 **公网地址**，形如：
   - 个人版：`registry.cn-hangzhou.aliyuncs.com`
   - 企业版：以控制台显示的实例域名称为准，通常是 `<实例名>-registry.<region>.cr.aliyuncs.com`

后续 GitHub Secret `ACR_REGISTRY` 就填这个地址（不要带 `https://`）。

### 1.2 创建 RAM 用户并授权

1. 进入 RAM 访问控制，创建用户 `github-actions-acr`，勾选「OpenAPI 调用访问」。
2. 授予权限：`AliyunContainerRegistryFullAccess`（或只包含推拉镜像权限的自定义策略）。
3. 保存 AccessKey（CI 里其实用不到 AK，但创建用户时需要）。
4. 在 ACR 的「访问凭证」页面为该 RAM 用户**设置固定密码**。这个密码就是 `ACR_PASSWORD`，不是 RAM 登录密码。
5. `ACR_USERNAME` 填 RAM 用户名，通常形如 `github-actions-acr@<账号别名>.onaliyun.com`，以控制台显示为准。

### 1.3 VPS 拉取权限

VPS 也需要登录 ACR 才能拉私有镜像。可以复用同一个 RAM 用户，或单独建一个只有拉取权限的账号。登录一次后凭据保存在 `~/.docker/config.json`：

```bash
docker login --username='<RAM用户名>' registry.cn-hangzhou.aliyuncs.com
```

如果使用企业版并绑定了 VPC：GitHub 托管 runner 只能用公网地址推送（`ACR_REGISTRY`），VPS 可以在 `.env` 里把 `ACR_IMAGE` 写成 VPC 地址以获得更快的拉取速度，但 VPS 也必须对该 VPC 地址执行一次 `docker login`。两个地址指向同一个镜像仓库，只是入口不同。

## 2. VPS 准备

以下以 Ubuntu 22.04/24.04 为例，2 核 2G、20G 磁盘足够。

### 2.1 安装 Docker

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo systemctl enable --now docker
docker --version
docker compose version
```

国内网络拉取 Docker Hub 镜像慢时，可在 `/etc/docker/daemon.json` 配置镜像加速，然后 `sudo systemctl restart docker`。

### 2.2 创建部署用户

```bash
sudo adduser --disabled-password --gecos "" deploy
sudo usermod -aG docker deploy
sudo mkdir -p /opt/onestarlab
sudo chown deploy:deploy /opt/onestarlab
```

> `deploy` 加入 `docker` 组等价于 root 权限，这是 Docker 的固有特性；VPS 只跑这一套服务时可以接受，务必只给可信的人密钥。

### 2.3 配置 SSH 密钥

在本地生成一对专用密钥（不要复用个人密钥）：

```bash
ssh-keygen -t ed25519 -C "github-actions-onestarlab" -f ~/.ssh/onestarlab_deploy
```

把公钥追加到 VPS 的 `deploy` 用户：

```bash
ssh-copy-id -i ~/.ssh/onestarlab_deploy.pub deploy@<VPS_IP>
```

可选加固：在 `/etc/ssh/sshd_config` 关闭密码登录（`PasswordAuthentication no`），并安装 fail2ban。

记录主机指纹，稍后作为 `VPS_KNOWN_HOSTS`：

```bash
ssh-keyscan -p 22 -H <VPS_IP>
```

### 2.4 准备部署目录

把仓库里的部署文件传到 VPS：

```bash
scp deploy/docker-compose.yml deploy/deploy.sh deploy/backup.sh deploy/.env.example \
  deploy@<VPS_IP>:/opt/onestarlab/
```

在 VPS 上：

```bash
cd /opt/onestarlab
cp .env.example .env
chmod +x deploy.sh backup.sh
sudo install -d -o 100 -g 101 data
```

`data` 目录必须属于容器内的 `uid=100(app)`、`gid=101(app)`，否则 SQLite 无法写入。

编辑 `.env`：

```dotenv
ACR_IMAGE=registry.cn-hangzhou.aliyuncs.com/onestarlab/onestarlab
IMAGE_TAG=latest
APP_PORT=8080
BIND_ADDR=0.0.0.0
TZ=Asia/Shanghai
API_TOKEN=
CORS_ORIGINS=*
```

说明：

- `ACR_IMAGE` 必须和 workflow 里的镜像名一致（ACR 地址 + 命名空间 + 仓库名）。
- `API_TOKEN` 建议先留空。前端 token 是构建时注入的，当前镜像没有注入；如果服务端设置了非空 token，页面会 401。需要访问控制时，优先用 HTTPS + 反向代理鉴权或安全组白名单。
- 如果前面要放 Caddy/Nginx，把 `BIND_ADDR` 改成 `127.0.0.1`，只让反代访问 8080。

### 2.5 首次登录 ACR 并部署

```bash
docker login --username='<RAM用户名>' registry.cn-hangzhou.aliyuncs.com
cd /opt/onestarlab
./deploy.sh latest
```

首次执行时 `.env` 里还没有上一版本，健康检查失败会直接报错退出；看日志排查即可。

### 2.6 防火墙与端口

```bash
sudo ufw allow OpenSSH
sudo ufw allow 8080/tcp     # 若走反向代理，这行可以不加
sudo ufw enable
sudo ufw status
```

云厂商的安全组也要放行对应端口。最终建议只暴露 `22`、`80`、`443`，8080 只对反向代理开放。

## 3. GitHub 配置

### 3.1 添加 Secrets

仓库 → Settings → Secrets and variables → Actions → New repository secret：

| Secret | 说明 | 示例 |
| --- | --- | --- |
| `ACR_REGISTRY` | ACR 公网地址，不带协议 | `registry.cn-hangzhou.aliyuncs.com` |
| `ACR_NAMESPACE` | ACR 命名空间 | `onestarlab` |
| `ACR_USERNAME` | RAM 用户名 | `github-actions-acr@xxx.onaliyun.com` |
| `ACR_PASSWORD` | ACR 访问凭证的固定密码 | `********` |
| `VPS_HOST` | VPS 公网 IP 或域名 | `47.xx.xx.xx` |
| `VPS_PORT` | SSH 端口，留空默认 22 | `22` |
| `VPS_USER` | 部署用户 | `deploy` |
| `VPS_SSH_KEY` | 私钥全文（含 BEGIN/END 行） | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `VPS_KNOWN_HOSTS` | `ssh-keyscan` 输出；为空则每次自动扫描 | `vps.example.com ssh-ed25519 AAAA...` |

可选 Repository variable：

| Variable | 说明 |
| --- | --- |
| `TARGET_PLATFORM` | 构建平台，默认 `linux/amd64`；ARM 机器改成 `linux/arm64` |

### 3.2 创建 production 环境（可选但推荐）

Settings → Environments → New environment → `production`。可以添加：

- Required reviewers（部署前人工批准）
- 部署分支限制（只允许 `main` 或 tag）

如果先不建环境，workflow 里的 `environment: production` 仍然能跑，但没有保护规则。

### 3.3 触发发布

```bash
git tag -a v1.0.0 -m "release v1.0.0"
git push origin v1.0.0
```

工作流依次执行：

1. `verify`：前端 `npm ci && npm run build`，后端 `go build ./... && go vet ./...`
2. `publish`：构建 `linux/amd64` 镜像，推送 `v1.0.0`、`latest` 和 buildcache
3. `deploy`：SSH 到 VPS，执行 `IMAGE_TAG=v1.0.0 ./deploy.sh`

也可以手动触发：Actions → Release → Run workflow；`tag` 输入留空时使用 `sha-<7位提交号>`。

## 4. 回滚

### 自动回滚

`deploy.sh` 在以下情况会尝试切回 `.env` 中记录的上一版本：

- `docker compose up -d` 失败
- 容器 60 秒内没有变成 `healthy`

回滚后脚本以非零状态退出，Actions 会显示红叉，需要人工看日志。

### 手动回滚

VPS 上直接执行旧 tag（镜像需仍在 ACR）：

```bash
cd /opt/onestarlab
./deploy.sh sha-abc1234
```

或在 GitHub Actions 里手动触发，`tag` 填旧版本号。

注意 ACR 若配置了保留策略，旧 tag 可能已被清理；重要版本建议保留。

## 5. HTTPS 与反向代理

最省事的是 Caddy：

```bash
sudo apt install -y caddy
sudo tee /etc/caddy/Caddyfile >/dev/null <<'EOF'
lab.example.com {
    reverse_proxy 127.0.0.1:8080
}
EOF
sudo systemctl reload caddy
```

同时把 `.env` 的 `BIND_ADDR` 改成 `127.0.0.1` 并重新部署，避免 8080 直接暴露公网。Caddy 会自动申请和续期证书。

如果继续用 8080 直连，至少确认：

- 云安全组只放行可信 IP；
- 不要把 `API_TOKEN` 当成真正的登录鉴权（token 会出现在浏览器请求里）。

## 6. 备份与恢复

### 自动备份

安装 `sqlite3`：

```bash
sudo apt install -y sqlite3
```

执行一次：

```bash
cd /opt/onestarlab
./backup.sh 14
```

加入 crontab（每天 03:00，保留 14 天）：

```cron
0 3 * * * /opt/onestarlab/backup.sh 14 >> /var/log/onestarlab-backup.log 2>&1
```

备份文件在 `/opt/onestarlab/backups/`，是 gzip 后的 SQLite 文件。建议再定期同步到对象存储或另一台机器。

### 恢复

```bash
cd /opt/onestarlab
docker compose stop
gunzip -c backups/onestar-YYYYMMDD-HHMMSS.db.gz > data/onestar.db
sudo chown 100:101 data/onestar.db
docker compose start
```

也可以继续用页面上的「数据中心 → 导入数据」恢复 JSON 快照，这种方式不需要停容器。

## 7. 常见问题

### `docker login` 失败或 push 401

- 确认 `ACR_REGISTRY` 与控制台一致（个人版/企业版地址不同）。
- 确认用的是 ACR 固定密码，不是阿里云登录密码。
- 确认 RAM 用户有 ACR 推拉权限。
- 在本地手工验证：

  ```bash
  docker login --username='<RAM用户名>' registry.cn-hangzhou.aliyuncs.com
  docker pull registry.cn-hangzhou.aliyuncs.com/<namespace>/onestarlab:latest
  ```

### Actions 的 deploy 步骤 SSH 失败

- `Permission denied`：`VPS_SSH_KEY` 是否是私钥全文、公钥是否在 `deploy` 的 `authorized_keys`。
- `Host key verification failed`：重新生成 `VPS_KNOWN_HOSTS`。
- 连接超时：安全组/防火墙未放行 SSH 端口，或 IP 填错。

### 容器一直 unhealthy

```bash
cd /opt/onestarlab
docker compose logs --tail 100 onestarlab
```

重点看：

- `/data` 权限是否 `100:101`；
- 8080 是否被占用；
- 数据库文件是否损坏。

### 拉取镜像报 `no matching manifest`

构建平台与 VPS 架构不一致。ARM 机器设置 `TARGET_PLATFORM=linux/arm64` 后重新发布。

### 页面能开但请求 401

服务端 `.env` 设置了 `API_TOKEN`，而前端构建时没有注入 `VITE_API_TOKEN`。把 `API_TOKEN` 清空并重新部署，或接受在前端注入 token（本质仍是公开的，不能替代登录）。

### 磁盘被旧镜像占满

```bash
docker image prune -a --filter "until=720h"
```

或给 ACR 配置保留策略，只保留最近若干个版本。

## 8. 安全检查清单

- [ ] ACR 仓库为私有，命名空间只给必要账号权限
- [ ] 使用独立 RAM 用户，遵循最小权限
- [ ] ACR 密码定期轮换，不写进仓库
- [ ] GitHub Secrets 只在仓库/环境保护中配置
- [ ] `production` 环境开启 Required reviewers
- [ ] VPS 禁用密码登录，仅密钥登录
- [ ] 安全组只放行 22/80/443，8080 不直接暴露
- [ ] 配置 HTTPS 后再对外使用
- [ ] 每日备份并验证可恢复
- [ ] 定期更新 VPS 系统与 Docker
- [ ] 关注磁盘、证书到期与容器健康状态

## 9. 本地可以提前验证的部分

```bash
# 生产编排语法与插值（需要先准备一个临时 env 文件）
docker compose -f deploy/docker-compose.yml --env-file deploy/.env.example config

# Shell 脚本语法（Git Bash / WSL）
bash -n deploy/deploy.sh
bash -n deploy/backup.sh

# 工作流语法可安装 actionlint 检查
actionlint .github/workflows/release.yml
```

镜像构建本身依赖 ACR 凭据，第一次真正跑通需要在 GitHub 上触发 release。
