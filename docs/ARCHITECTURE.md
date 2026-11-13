# OneStarLab 架构说明

## 总览

OneStarLab 是单体前端 SPA，采用 Vue 3 + Pinia + Vue Router + Dexie。

核心数据流：

View -> Store -> Dexie Table

设计目标：

- 在无后端场景下，保证店铺管理功能可本地离线运行。
- 通过 Store 统一业务流程，避免页面直接操作数据库。
- 将业务类型与计算逻辑集中，降低维护成本。

## 分层与职责

### 视图层

目录：[src/views](src/views)

职责：

- 展示页面与交互。
- 表单输入校验与用户反馈。
- 调用 Store action，不直接调用 Dexie。

### 状态与业务编排层

目录：[src/stores/useShopStore.ts](src/stores/useShopStore.ts)

职责：

- 管理 products、orders、expenses、markets、pricingRule、labInspirations、labProjects、plannerTemplates。
- 维护线上收入（`settings` 表中 `onlineIncome` 键）与更新时间，收入页据此计算总收入与利润。
- 提供 CRUD action 与业务流程（例如下单扣减库存）。
- 通过 loadAll 统一刷新状态，确保 UI 与数据库一致。

关键约束：

- 同时修改订单与库存时，必须使用 Dexie transaction。
- 对外暴露明确错误信息，便于页面提示。

### 数据访问层

目录：[src/db/appDb.ts](src/db/appDb.ts)

职责：

- 定义 IndexedDB 数据库和版本。
- 定义表结构与索引字段。

版本策略：

- 变更 schema 时必须增加 version。
- 为常用查询字段建立索引（例如 createdAt、status、tag、purpose、date、location）。

### 组件层

目录：[src/components/planner](src/components/planner)、[src/components/tape](src/components/tape)

职责：

- `PlannerSheet.vue` 按毫米尺寸渲染一张内页，负责留白、模块间距与 CSS 变量。
- `PlannerBlock.vue` 负责 15 种内页模块的模板与样式。
- `TapeCanvas.vue` 接收 `TapeDesign`，按容器宽度换算预览比例后交给 canvas 重绘。
- 组件只接收 props 并渲染，不读写 store，也不感知打印或下载流程。

关键约束：

- 纸张与模块尺寸统一使用 mm，字号用 em，避免 px 硬编码导致打印比例失真。
- 屏幕预览与打印共用同一套组件，打印通过 `@page` 与 `@media print` 切换，不另做一套排版。

### 类型层

目录：[src/types/index.ts](src/types/index.ts)

职责：

- 统一管理 Product、Order、PricingRule、LabProject、MarketRecord、AppSetting、PlannerTemplate、TapeDesign 等业务模型。
- 作为 Store、View、DB 的共同契约。

建议：

- 新增字段时先更新类型，再实现业务逻辑和界面。

### 工具层

目录：[src/utils/pricing.ts](src/utils/pricing.ts)、[src/utils/tape.ts](src/utils/tape.ts)

职责：

- `pricing.ts` 存放定价相关的纯函数计算逻辑（如建议售价、最小售价）。
- `tape.ts` 负责胶带图案绘制、毫米到像素换算、PNG/PSD 导出，视图层只传设计参数与 DPI。
- 避免将公式或绘制细节分散在页面组件中。

## 路由设计

路由定义见：[src/router/index.ts](src/router/index.ts)

特性：

- 页面组件采用懒加载。
- 使用 meta.title 并在 afterEach 中统一更新 document.title。
- 收入记录统一在 `/income` 页面（集市收入 + 线上收入 + 利润），旧的 `/markets` 保留为跳转到 `/income` 的重定向。
- 印章相关能力统一在单一路由页面中，通过页内 Tab 切换文字印章与点阵印章。
- 内页生成器位于 `/planner`，模板持久化在 `plannerTemplates`，打印时动态注入 `@page { size: Wmm Hmm; margin: 0 }` 以匹配所选尺寸。
- 胶带生成位于 `/tape`，设计只存在页面状态，不写 IndexedDB；导出由 `src/utils/tape.ts` 按 DPI 生成 canvas。

## 第三方库使用要点

- ECharts：在组件卸载时释放实例，防止内存泄漏。
- Dexie：跨表写操作使用 transaction，确保一致性。
- Pinia：保持 action 粒度清晰，避免页面中重复业务逻辑。
- ag-psd：只在导出 PSD 时动态 import，避免增加首屏体积；导出后按图层名组织图案。

## 当前架构边界

- 无后端 API 与账号系统。
- 无远程同步能力。
- 数据生命周期依赖浏览器本地存储策略。

如需演进为多端同步架构，建议优先拆分：

1. 领域模型（types）
2. 业务服务层（store 可逐步迁移到 service）
3. 存储适配层（Dexie / HTTP API）
