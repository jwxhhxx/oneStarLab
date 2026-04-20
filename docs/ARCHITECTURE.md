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

- 管理 products、orders、expenses、pricingRule、labInspirations、labProjects。
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
- 为常用查询字段建立索引（例如 createdAt、status、tag、purpose）。

### 类型层

目录：[src/types/index.ts](src/types/index.ts)

职责：

- 统一管理 Product、Order、PricingRule、LabProject 等业务模型。
- 作为 Store、View、DB 的共同契约。

建议：

- 新增字段时先更新类型，再实现业务逻辑和界面。

### 工具层

目录：[src/utils/pricing.ts](src/utils/pricing.ts)

职责：

- 存放纯函数计算逻辑（如建议售价、最小售价）。
- 避免将公式分散在页面组件中。

## 路由设计

路由定义见：[src/router/index.ts](src/router/index.ts)

特性：

- 页面组件采用懒加载。
- 使用 meta.title 并在 afterEach 中统一更新 document.title。

## 第三方库使用要点

- ECharts：在组件卸载时释放实例，防止内存泄漏。
- Dexie：跨表写操作使用 transaction，确保一致性。
- Pinia：保持 action 粒度清晰，避免页面中重复业务逻辑。

## 当前架构边界

- 无后端 API 与账号系统。
- 无远程同步能力。
- 数据生命周期依赖浏览器本地存储策略。

如需演进为多端同步架构，建议优先拆分：

1. 领域模型（types）
2. 业务服务层（store 可逐步迁移到 service）
3. 存储适配层（Dexie / HTTP API）
