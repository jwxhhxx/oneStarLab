# OneStarLab 开发与维护指南

## 环境要求

- Node.js 18+
- npm 9+

## 常用命令

- 安装依赖：npm install
- 本地开发：npm run dev
- 生产构建（含类型检查）：npm run build
- 产物预览：npm run preview

说明：当前未配置独立测试与 lint 命令。

## 推荐开发流程

1. 先更新业务类型：[src/types/index.ts](src/types/index.ts)
2. 再补齐数据库结构：[src/db/appDb.ts](src/db/appDb.ts)
3. 在 Store 中实现业务流程：[src/stores/useShopStore.ts](src/stores/useShopStore.ts)
4. 最后更新页面交互：[src/views](src/views)

## 关键约定

- 不在视图层直接操作 Dexie。
- 定价相关公式放在 [src/utils/pricing.ts](src/utils/pricing.ts) 纯函数中。
- 涉及订单和库存同时变更的逻辑必须使用 transaction。
- 写入数据库后通过 loadAll 同步状态。
- 页面新增文案保持中文语境一致。
- 支出记录属于独立业务数据，新增字段时同步更新支出页面、store 和 Dexie schema。

## IndexedDB Schema 变更清单

当你修改数据库结构时：

1. 在 [src/db/appDb.ts](src/db/appDb.ts) 增加 Dexie version。
2. 调整 stores 字段与索引。
3. 同步更新类型定义 [src/types/index.ts](src/types/index.ts)。
4. 检查 Store 读写逻辑是否兼容新字段。
5. 运行 npm run build 验证类型与构建。

## 常见问题

### 修改了数据但页面未刷新

优先检查对应 action 是否在写库后调用了 loadAll。

### 下单后库存异常

检查：

- 是否在 transaction 中同时写 orders 与 products。
- updateOrder 场景是否正确回补旧商品库存。

### 页面标题不正确

检查 [src/router/index.ts](src/router/index.ts) 中 route meta.title 与 afterEach 逻辑。

### 图表重复渲染或内存占用增长

检查图表组件是否在 onUnmounted 时调用 dispose。

## 发布前检查

1. npm run build 成功。
2. 核心路径手动验证：商品新增、订单新增、订单编辑、订单删除、定价规则保存。
3. 验证支出记录页的新增/编辑/删除流程。
4. 验证研究所与灵感页的新增/编辑/删除流程。
5. 确认浏览器刷新后本地数据仍可读。
