# OneStarLab 开发与维护指南

## 环境要求

- Node.js 18+、npm 9+
- Go 1.22+
- 可选：Docker 24+ 与 Docker Compose v2（容器运行/验收）

Windows 上如果 PowerShell 执行策略禁止 `npm.ps1`，把命令里的 `npm` 换成 `npm.cmd`。

## 常用命令

前端：

- 安装依赖：`npm ci`（或 `npm install`）
- 本地开发：`npm run dev`（默认 5173，`/api` 代理到 8080）
- 类型检查 + 生产构建：`npm run build`
- 产物预览：`npm run preview`（只预览静态文件，API 仍需单独启动）

后端（在 `backend/` 下）：

- 直接运行：`go run ./cmd/server`
- 编译：`go build ./...`
- 静态检查：`go vet ./...`
- 全量检查：`go build ./... && go vet ./...`

容器：

- 构建并启动：`docker compose up --build -d`
- 查看状态：`docker compose ps`、`docker compose logs -f`
- 停止：`docker compose down`（`./data` 中的数据库保留）

生产发布（需要先完成 [DEPLOYMENT.md](DEPLOYMENT.md) 里的 ACR / VPS / Secrets 配置）：

- 打标签发布：`git tag -a v1.0.0 -m "release v1.0.0" && git push origin v1.0.0`
- 手动触发：GitHub Actions → Release → Run workflow（可勾选 `skip_deploy` 只推镜像）
- 仓库不包含 VPS 侧的 compose 与脚本；部署命令在 `VPS_APP_DIR` 下执行 `docker compose pull && up -d`

说明：当前没有单元测试脚本；后端接口与页面流程按下方回归清单人工验证。

## 推荐开发流程

改动一个业务功能时，按下面的顺序推进：

1. 更新类型契约：[src/types/index.ts](src/types/index.ts) 与 [backend/internal/model/model.go](backend/internal/model/model.go)。
2. 改数据库结构（如需要）：[backend/internal/db/db.go](backend/internal/db/db.go) 的 `migrate`。
3. 实现后端业务与 SQL：[backend/internal/store](backend/internal/store)。
4. 暴露 HTTP 接口：[backend/internal/api/api.go](backend/internal/api/api.go) 加路由，对应 `handlers_*.go` 加 handler。
5. 更新前端 API 客户端：[src/api/client.ts](src/api/client.ts)。
6. 在 Store 中接入：[src/stores/useShopStore.ts](src/stores/useShopStore.ts)。
7. 最后改页面交互：[src/views](src/views)。
8. 运行 `go build ./... && go vet ./...` 与 `npm run build`。

## 关键约定

- 视图层不直接调用 `src/api/client.ts`，统一经过 store。
- 后端返回中文错误信息，前端直接展示；不要在前端重复拼接同类文案。
- 金额统一保留两位小数；定价用后端 `roundPrice`，前端 `src/utils/pricing.ts` 只用于表单即时预览。
- 同时修改订单与库存、批量扫码、快照导入等多表操作必须放在 SQL 事务中。
- 写入成功后前端调用 `loadAll()` 重新拉取 `bootstrap`，保持 UI 与服务端一致。
- 页面文案保持中文语境一致。
- 支出记录、集市收入、线上收入是独立业务数据，新增字段时同步更新模型、SQL、API、store 与页面。
- 内页生成器的纸张与模块尺寸统一用 mm，字号用 em；屏幕预览与打印共用 [src/components/planner](src/components/planner) 下的同一套组件。
- 胶带绘制与导出逻辑放在 [src/utils/tape.ts](src/utils/tape.ts)，尺寸以 mm 为单位、按 DPI 换算像素；PSD 依赖 `ag-psd` 只在导出时动态 import。
- 认证是“单口令 + 会话 Cookie”：新增公开接口必须显式加入 [backend/internal/api/auth.go](backend/internal/api/auth.go) 的公开路径列表；其余接口默认要求登录，开发时不要用 `API_TOKEN` 绕过验证。

## 数据库结构变更清单

1. 在 [backend/internal/db/db.go](backend/internal/db/db.go) 的 `migrate` 中增加 `CREATE TABLE` / `ALTER TABLE`（当前项目用幂等建表，尚无版本号迁移框架）。
2. 同步更新 [backend/internal/model/model.go](backend/internal/model/model.go) 的字段与 JSON tag。
3. 更新 `internal/store` 的查询与扫描逻辑（`SELECT` 列、`Scan` 顺序）。
4. 更新前端类型与 API 客户端。
5. 旧库升级需要保证幂等：新增列时用 `ALTER TABLE ... ADD COLUMN`，并在启动时容错判断列是否已存在。
6. 用 `go build ./...`、`npm run build` 与手工回归验证。

## 常见问题

### 前端页面显示「无法连接后端服务」

- 确认后端已启动：`curl http://127.0.0.1:8080/api/v1/health`。
- 确认 Vite 代理目标正确：默认 `http://127.0.0.1:8080`，可用 `VITE_PROXY_TARGET` 覆盖。
- 8080 被旧进程或容器占用时，先停止占用者。

### 修改了数据但页面未刷新

检查对应 action 是否在写操作成功后调用了 `loadAll()`。

### 下单后库存异常

检查：

- 后端是否在同一事务中写 `orders`、`products` 与 `inventory_transactions`。
- 编辑订单时是否回补旧商品库存（当前沿用「只处理第一件商品」的既有语义）。

### 登录后仍然跳回登录页

- 检查 Cookie 是否被浏览器拒绝：跨域部署时前端 `VITE_API_BASE` 必须与后端同站，或后端 `CORS_ORIGINS` 明确允许来源；跨站 Cookie 需要 `SameSite=None; Secure`，本项目默认按同站部署设计。
- 通过 HTTPS 反向代理时，确认代理传递 `X-Forwarded-Proto: https`，否则 Cookie 不带 `Secure` 不影响登录，但安全属性会降级。
- 修改口令后旧 Cookie 会失效，必须重新登录。

### 忘记口令

停掉服务后清空 `settings` 表中的认证键并重启，系统会重新写入默认口令 `123456`：

```bash
sqlite3 data/onestar.db "DELETE FROM settings WHERE key LIKE 'auth.%';"
```

### 容器启动失败或拉取基础镜像超时

本机访问 Docker Hub 可能超时。可先从镜像站拉取并打本地 tag，再构建：

```bash
docker pull docker.m.daocloud.io/library/node:22-alpine
docker tag docker.m.daocloud.io/library/node:22-alpine node:22-alpine
docker pull docker.m.daocloud.io/library/golang:1.24-alpine
docker tag docker.m.daocloud.io/library/golang:1.24-alpine golang:1.24-alpine
docker compose build
```

`Dockerfile` 顶部不要加 `# syntax=docker/dockerfile:1`，该指令会额外拉取 BuildKit frontend 镜像。

### 中文在 PowerShell 对比时“丢失”

`Invoke-RestMethod` 的响应经控制台编码后中文可能显示为 `?`，用 `Where-Object` 比较会得到假阴性。请用 `curl.exe` 查看原始 JSON，或按 `id` 比对。

### 页面标题不正确

检查 [src/router/index.ts](src/router/index.ts) 的 `meta.title` 与 `router.afterEach`。

### 修改代码后 dev 页面未实时更新

1. 强制刷新（Ctrl+Shift+R）。
2. 确认只有一个 dev 进程且端口一致。
3. 重启 `npm run dev`。

## 发布前检查

1. `go build ./... && go vet ./...` 与 `npm run build` 均成功。
2. 登录：未登录访问任意内页会跳转登录页；默认口令 `123456` 可登录；错误口令提示“口令不正确”；连续失败会触发 60 秒锁定。
3. 设置：修改口令后旧口令失效、新口令可登录；其他浏览器旧会话被踢回登录页。
4. 核心路径：商品新增、订单新增/编辑/删除、定价规则保存。
5. 支出记录新增/编辑/删除。
6. 收入记录：集市新增/编辑/删除、线上收入保存、总收入与利润计算、负利润标红。
7. 内页生成器：尺寸切换、模块增删与参数调整、模板保存/载入、打印纸张尺寸与所选 mm 一致。
8. 胶带生成：宽长切换、图案层增删与排数/列数/大小调整、透明底、PNG 与 PSD 导出的尺寸和图层结构。
9. 研究所与灵感页新增/编辑/删除。
10. 刷新浏览器后数据仍在（服务端 SQLite）；退出登录后再访问会回到登录页。
11. 数据中心导出 JSON，再导入同一文件，数据保持一致。
12. `docker compose up --build -d` 后 `docker compose ps` 显示 healthy，页面与 API 均可访问；重启容器后数据不丢。
13. 发布生产前：`release.yml` 三段（verify / publish / deploy）在测试 tag 上跑通，deploy 末尾的容器健康检查通过。
