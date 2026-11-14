# OneStarLab 手帐店铺助手

OneStarLab 是前后端分离的手帐店铺经营后台：Vue 3 单页应用 + Go REST API + 单个 SQLite 数据库。

浏览器只负责界面，所有业务数据保存在服务端 SQLite 中；不再依赖 GitHub Pages，也不再使用 IndexedDB。

## 功能概览

当前版本支持：

- 商品与 SKU 管理
- 库存追踪与低库存预警
- 订单录入与利润核算
- 开店支出记录与累计统计
- 收入台账：集市按场记录，线上收入手工定期更新
- 收入合计 − 支出合计自动算出利润，负数标红提醒
- 自动定价建议
- 基础经营看板
- 研发流程管理（研究所）
- 灵感随机生成功能
- 印章生成（同页支持文字印章与点阵印章）
- 内页生成器：按毫米拼装手帐内页，支持打印和导出 1:1 PDF
- 胶带生成：自定义胶带宽长、图案排布与配色，导出 PNG 或分层 PSD
- 数据中心：导出 / 导入整库 JSON 快照

## 页面说明

- 经营看板：查看销售额、利润、订单趋势等核心指标
- 商品中心：维护商品、库存、安全库存与建议售价
- 订单中心：录入与维护订单，自动计算成本和净利润
- 支出记录：记录开店以来的支出用途与金额，形成成本台账
- 收入记录：按场次记录集市名称、日期、地点与收入，线上收入手工维护；页面汇总总收入与支出并计算利润，利润为负时标红提醒
- 自动定价：维护目标毛利率、平台费率和取整策略
- 研究所：管理灵感到上新的项目进度
- 灵感生成：按品类抽取命名、风格、主题等灵感
- 印章生成：同一页面内通过 Tab 切换文字印章与点阵印章
- 内页生成器：选择 M5、A7、A6per、三折页等尺寸，组合月历、时间轴、待办、四象限、账单、格纸等模块，调整配色与打印份数后直接打印或导出 PDF
- 胶带生成：自定义胶带宽度与长度，叠加波点、星星、爱心、钥匙、翅膀等图案层，逐层设置排数、列数、大小、旋转与配色，导出 PNG 或可继续编辑的分层 PSD
- 数据中心：显示服务端存储状态，导出整库 JSON，或选择备份文件覆盖导入

## 技术栈

前端：

- Vue 3 + TypeScript
- Vite
- Pinia
- Element Plus
- ECharts
- ag-psd（胶带 PSD 导出，按需加载）

后端：

- Go 1.22+，标准库 `net/http`（Go 1.22 ServeMux 方法路由）
- SQLite：`modernc.org/sqlite`（纯 Go 实现，无需 CGO）
- REST JSON API，统一前缀 `/api/v1`

部署：

- Docker 多阶段构建（Node 构建前端 → Go 编译后端 → Alpine 运行）
- Docker Compose 单服务编排，SQLite 通过 volume 持久化

## 项目结构

```text
src/                    前端
  api/client.ts         统一 REST 客户端
  components/           planner / tape 可复用组件
  router/               路由与页面标题
  stores/useShopStore.ts 业务编排层（调用 API、维护页面状态）
  types/                业务类型
  utils/                纯函数工具（定价、胶带绘制）
  views/                页面视图
backend/
  cmd/server/           服务入口
  internal/api/         路由、中间件与 HTTP handler
  internal/store/       业务逻辑与 SQL
  internal/db/          建表、迁移与种子数据
  internal/model/       请求/响应结构体
docs/                   架构、业务、开发、API 与验收文档
Dockerfile
compose.yaml
deploy/                 生产部署（ACR 镜像编排、deploy.sh、backup.sh）
data/                   Compose 挂载的 SQLite 目录（已 gitignore）
```

## 本地开发

前置要求：

- Node.js 18+ 与 npm
- Go 1.22+

```bash
# 终端 1：启动后端（默认监听 8080）
cd backend
go run ./cmd/server

# 终端 2：启动前端（默认 5173，/api 代理到 8080）
npm ci
npm run dev
```

访问 <http://localhost:5173>。前端默认请求同源 `/api/v1`，由 Vite 代理到 `http://127.0.0.1:8080`；后端地址不同时设置 `VITE_PROXY_TARGET`。

健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/health
```

后端默认数据库是 `backend/data/onestar.db`（相对当前工作目录），可用 `DATABASE_PATH` 指定。

## Docker 运行

```bash
docker compose up --build -d
```

- 应用地址：<http://localhost:8080>
- 数据库文件：宿主机 `./data/onestar.db`，容器内 `/data/onestar.db`
- 查看日志：`docker compose logs -f`
- 停止：`docker compose down`（SQLite 文件保留在 `./data`）

容器内置健康检查（`/api/v1/health`），`docker compose ps` 显示 `healthy` 即表示服务与静态资源都正常。

## 生产部署

生产方案：GitHub Actions 构建镜像 → 推送阿里云 ACR → Actions 通过 SSH 在 VPS 上执行 `docker compose pull && up -d`。

仓库已包含：

- [`.github/workflows/release.yml`](.github/workflows/release.yml)：tag 发布、构建推送、SSH 部署
- [`deploy/docker-compose.yml`](deploy/docker-compose.yml)：VPS 侧编排（使用 ACR 镜像，加入 external `proxy` 网络，由 NPM 等反代，不映射宿主机端口）
- [`deploy/deploy.sh`](deploy/deploy.sh)：拉取、启动、健康检查与自动回滚
- [`deploy/backup.sh`](deploy/backup.sh)：SQLite 备份

阿里云 ACR、VPS 与 GitHub Secrets 的完整配置步骤见 [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)。

## 构建与检查

```bash
npm run build            # vue-tsc 类型检查 + Vite 生产构建
cd backend && go build ./... && go vet ./...
docker compose build     # 构建运行镜像
```

当前没有单元测试脚本；发布前按 [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) 的人工回归清单执行。

## 环境变量

后端：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | HTTP 监听端口 |
| `DATABASE_PATH` | `data/onestar.db` | SQLite 文件路径，容器内为 `/data/onestar.db` |
| `WEB_DIR` | 空 | 前端静态目录；为空时只提供 API |
| `API_TOKEN` | 空 | 设置后所有 `/api/v1` 请求需携带 `Authorization: Bearer <token>` |
| `CORS_ORIGINS` | `*` | 允许的来源，逗号分隔；`*` 表示不限制 |

前端（构建时注入）：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `VITE_API_BASE` | `/api/v1` | API 前缀，前后端不同域时改成完整地址 |
| `VITE_API_TOKEN` | 空 | 与后端 `API_TOKEN` 对应 |
| `VITE_PROXY_TARGET` | `http://127.0.0.1:8080` | 仅开发期 Vite 代理目标 |

## 数据与备份

- 全部业务数据保存在服务端 SQLite，浏览器不再持久化业务数据。
- 备份方式：数据中心「导出备份」下载 JSON 快照，或直接复制 `data/onestar.db`。
- 导入会覆盖服务端全部业务数据，导入前先导出备份。
- 浏览器里的旧版 IndexedDB 数据不会自动迁移；但旧版纯前端版本在「数据中心」导出的 JSON 备份可以直接在新版「数据中心」导入，服务端会自动转换为新版快照（支持商品、订单、分类、定价规则、研究所、支出、集市、线上收入、内页模板等全部旧表）。

## 相关文档

- 架构与数据流：[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- 业务说明与操作流程：[docs/BUSINESS.md](docs/BUSINESS.md)
- 开发与维护指南：[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
- 生产部署（ACR + VPS）：[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)
- REST API 参考：[docs/API.md](docs/API.md)
- 本地验收清单：[docs/ACCEPTANCE.md](docs/ACCEPTANCE.md)
