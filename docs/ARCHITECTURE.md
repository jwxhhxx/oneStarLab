# OneStarLab 架构说明

## 总览

OneStarLab 采用前后端分离结构：

- 前端：Vue 3 单页应用（SPA），负责界面、交互与展示状态。
- 后端：Go HTTP 服务，提供 `/api/v1` REST API，负责全部业务规则与持久化。
- 数据库：单个 SQLite 文件（默认 `data/onestar.db`）。

核心数据流：

```text
View → Pinia store → src/api/client.ts → HTTP /api/v1
     → internal/api（路由/handler）→ internal/store（业务与事务）
     → database/sql → SQLite
```

设计目标：

- 单一二进制部署：Go 服务同时提供 API 与构建后的前端静态文件。
- 业务规则集中在后端 store 层，页面不直接拼 SQL、不重复实现计算口径。
- 前端保留原有组件与样式，只把数据访问从 IndexedDB 换成 REST。
- SQLite 足够单店单机使用，备份就是一个文件。

## 分层与职责

### 前端视图层

目录：[src/views](src/views)

职责：

- 页面展示、表单校验、用户反馈。
- 调用 store action，不直接发 HTTP 请求。

### 前端状态与业务编排层

目录：[src/stores/useShopStore.ts](src/stores/useShopStore.ts)

职责：

- 维护 products、orders、expenses、markets、pricingRule、labInspirations、labProjects、plannerTemplates 等响应式状态。
- 把页面动作翻译成 API 调用，写操作完成后调用 `loadAll()` 重新拉取 `bootstrap`。
- 保留看板所需的计算属性（`stats`、`monthlyTrend`、`upcomingLaunches`、`lowStockProducts`）。
- 维护线上收入与更新时间，供收入页计算总收入与利润。

关键约束：

- 视图层不得直接 import `src/api/client.ts` 发请求，统一走 store。
- 金额、利润、库存等口径由后端计算；前端只做展示与轻量本地汇总。
- 对外暴露明确错误信息，便于页面提示。

### 前端 API 客户端

目录：[src/api/client.ts](src/api/client.ts)

职责：

- 统一 base URL（默认 `/api/v1`）、JSON 序列化、Bearer Token（可选）与 Cookie 会话（`credentials: 'include'`）。
- 把非 2xx 响应的 `{ "error": "..." }` 转成 `Error` 抛出；收到 401 时广播 `onestarlab:unauthorized`，由 `main.ts` 统一跳转登录页。
- 为每类资源提供方法：auth、products、categories、orders、pricing、inventory、lab、drawings、expenses、markets、settings、planner、export/import。

### 前端组件层

目录：[src/components/planner](src/components/planner)、[src/components/tape](src/components/tape)

职责：

- `PlannerSheet.vue` 按毫米尺寸渲染一张内页，负责留白、模块间距与 CSS 变量。
- `PlannerBlock.vue` 负责 15 种内页模块的模板与样式。
- `TapeCanvas.vue` 接收 `TapeDesign`，按容器宽度换算预览比例后交给 canvas 重绘。
- 组件只接收 props 并渲染，不读写 store，也不感知打印或下载流程。

关键约束：

- 纸张与模块尺寸统一使用 mm，字号用 em，避免 px 硬编码导致打印比例失真。
- 屏幕预览与打印共用同一套组件，打印通过 `@page` 与 `@media print` 切换，不另做一套排版。

### 前端类型层

目录：[src/types/index.ts](src/types/index.ts)

职责：

- 统一管理 Product、Order、PricingRule、LabProject、MarketRecord、AppSetting、PlannerTemplate、TapeDesign、BootstrapData 等模型。
- 作为 Store、View、API 客户端的共同契约。

### 前端工具层

目录：[src/utils/pricing.ts](src/utils/pricing.ts)、[src/utils/tape.ts](src/utils/tape.ts)

职责：

- `pricing.ts`：建议售价、最低售价与金额格式化等纯函数。
- `tape.ts`：胶带图案绘制、毫米到像素换算、PNG/PSD 导出。
- 公式与绘制细节不分散到页面组件中。

### 后端入口层

目录：[backend/cmd/server](backend/cmd/server)

职责：

- 读取环境变量（`PORT`、`DATABASE_PATH`、`WEB_DIR`、`API_TOKEN`、`CORS_ORIGINS`）。
- 打开数据库、装配 `api.NewServer(...)`、启动/优雅关闭 HTTP 服务。

### 后端 HTTP 层

目录：[backend/internal/api](backend/internal/api)

职责：

- Go 1.22 ServeMux 方法路由，统一前缀 `/api/v1`。
- 认证：公开路径仅 `health` 与 `/auth/*`；其余请求接受会话 Cookie 或可选 `API_TOKEN` Bearer，否则 401。
- 登录限流（同一 IP 连续失败 5 次锁 60 秒）、会话 Cookie 签发/清理、口令修改后轮换会话密钥。
- CORS、请求日志、JSON 编解码（限制 64MB、拒绝未知字段）。
- 错误映射：`ErrNotFound` → 404；常见业务校验错误 → 400；其余 → 500。
- 静态资源：从 `WEB_DIR` 提供构建产物，未命中的路径回退 `index.html`。

### 后端认证与数据文件

目录：[backend/internal/auth](backend/internal/auth)

职责：

- 口令哈希：PBKDF2-HMAC-SHA256（12 万次迭代、16 字节随机盐），比较使用常量时间。
- 会话：HMAC-SHA256 签名的 `payload.sig` 令牌，payload 含过期时间与随机 nonce；默认 7 天。
- 口令与盐、会话密钥存放在 `settings` 表的 `auth.*` 键中；首次启动写入默认口令 `123456` 并标记 `auth.passcodeIsDefault=1`。

### 后端业务层

目录：[backend/internal/store](backend/internal/store)

职责：

- 全部业务规则：定价取整、订单金额与利润、库存变更、扫码建单、研发阶段进度、SKU 生成、快照导入导出。
- 使用 `database/sql` 与手写 SQL；多表联动放在同一事务中。
- 对上层返回领域错误（如 `商品 %s 库存不足`、`默认分类“未分类”不可删除`）。

### 后端数据层

目录：[backend/internal/db](backend/internal/db)

职责：

- 建表与索引（`migrate`）。
- 种子数据：默认定价规则、`未分类` 与四个默认分类（`INSERT OR IGNORE`，可重复执行）。
- 认证种子：首次启动写入默认口令哈希、随机会话密钥与 `auth.passcodeIsDefault=1`，重启不覆盖已有值。

### 后端模型层

目录：[backend/internal/model](backend/internal/model)

职责：

- 请求/响应结构体，JSON tag 与前端类型字段一致（camelCase）。
- `Bootstrap`（前端一次拉取的全部数据）与 `Snapshot`（导入导出格式，`version: 1`）。

## 数据存储

- SQLite 使用 `modernc.org/sqlite`（纯 Go，无 CGO）。
- 连接池限制为单连接（`SetMaxOpenConns(1)`），配合 `journal_mode=WAL`、`synchronous=NORMAL`、`busy_timeout=5000`，避免并发写冲突。
- 订单商品项、研发打样记录、内页模板模块以 JSON 文本列存储；金额用 REAL，时间用 UTC ISO 字符串（毫秒精度）。
- 业务表共 12 张：`products`、`categories`、`orders`、`pricing_rules`、`lab_inspirations`、`lab_projects`、`drawing_inspirations`、`expenses`、`inventory_transactions`、`markets`、`settings`、`planner_templates`。
- `settings` 同时保存线上收入（`onlineIncome`）与认证数据（`auth.passcodeHash`、`auth.passcodeSalt`、`auth.passcodeIterations`、`auth.passcodeIsDefault`、`auth.sessionSecret`）。

### 一致性与事务

- 创建/编辑/删除订单与库存回滚在同一事务内完成。
- 批量扫码（`/inventory/scan`）在同一事务内更新库存、写流水并按扫码结果建单。
- 导入快照时先清空全部业务表，再按原 id 写入并修正自增序列，全程单事务，失败即回滚。
- 旧版纯前端备份由 `internal/store/legacy.go` 解析 `dexie-export-import` 格式，转换为同一套 `Snapshot` 后再走上面的导入事务；两个入口共用同一覆盖语义。

## API 约定

- 路径前缀 `/api/v1`，全部 JSON。
- 成功返回资源对象、数组或 `{ "ok": true }`；失败返回 `{ "error": "中文错误信息" }` 与非 2xx 状态码。
- 列表接口不分页（单店数据量小），一次返回全量。
- `GET /bootstrap` 一次返回首页所需全部数据，避免前端多次往返。
- `GET /export` 返回附件 `onestarlab-backup.json`；`POST /import` 覆盖导入。
- 认证使用登录 Cookie；设置 `API_TOKEN` 后，非公开路径也可用 `Authorization: Bearer <token>`，适合脚本调用。公开路径为 `health` 与 `/auth/*`。

接口清单见 [docs/API.md](API.md)。

## 前端路由

路由定义见：[src/router/index.ts](src/router/index.ts)

特性：

- 使用 hash 路由（`createWebHashHistory`），静态托管无需服务端 rewrite。
- 全局守卫先调用 `useAuthStore().check()`：未登录时除 `/login` 外都跳转到登录页并记录 `redirect`；已登录访问 `/login` 则回首页。
- 页面组件懒加载，`meta.title` 在 `router.afterEach` 统一写入 `document.title`。
- 收入记录统一在 `/income`（集市收入 + 线上收入 + 利润），旧 `/markets` 重定向过去。
- 印章能力集中在单一路由，页内 Tab 切换文字与点阵印章。
- 内页生成器位于 `/planner`，模板保存到服务端 `planner_templates`；打印时动态注入 `@page { size: Wmm Hmm; margin: 0 }`。
- 胶带生成位于 `/tape`，设计只存在页面状态，不写数据库。
- 设置页位于 `/settings`：修改访问口令，成功后会提示下次使用新口令。

## 构建与部署

- 前端：`npm run build` 输出 `dist/`（含 vue-tsc 类型检查）。
- 后端：`go build ./...`；容器内 `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"`。
- 镜像（[Dockerfile](../Dockerfile)）三阶段：Node 构建前端 → Go 编译后端 → Alpine 运行；运行阶段非 root 用户，`WEB_DIR=/app/web`，`/data` 为数据卷。
- 编排（[compose.yaml](../compose.yaml)）：本地单服务，`8080:8080`，`./data:/data`，内置健康检查。
- 生产发布（[release.yml](../.github/workflows/release.yml)）：tag 触发 → 构建并推送阿里云 ACR → SSH 到 VPS，在配置好的目录执行 `docker compose pull && up -d`；VPS 侧的 compose 与数据由部署方维护，仓库不保存路径、用户等环境配置。
- 不再使用 GitHub Pages；`ci.yml` 只做构建与校验，`release.yml` 负责镜像与部署。

## 第三方库使用要点

- ECharts：组件卸载时释放实例，防止内存泄漏。
- Element Plus：全量引入，构建体积偏大；后续可按需引入优化。
- ag-psd：只在导出 PSD 时动态 `import()`，不进入首屏包。
- Pinia：action 粒度保持清晰，页面不重复业务逻辑。

## 当前架构边界

- 单用户、单实例：没有多账号体系；只有一道访问口令与可选 `API_TOKEN`，适合个人自用。
- 没有实时同步或多端推送；多设备同时操作时以最后一次写入为准。
- 登录会话 7 天有效；修改口令会轮换密钥并使其他旧会话失效。
- 没有分页与全文搜索，适合单店数据规模。
- 旧纯前端版本的 IndexedDB 数据不会自动迁移到 SQLite。

后续演进方向：

1. 多用户/多店与细粒度权限。
2. 订单、商品的查询分页与筛选。
3. 定时备份与远端对象存储。
4. 需要实时协作时再引入 WebSocket/SSE 与服务端版本号。
