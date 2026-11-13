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
- 支出记录、集市收入、线上收入配置属于独立业务数据，新增字段时同步更新对应页面、store 和 Dexie schema。
- 内页生成器的纸张与模块尺寸统一用 mm，字号用 em；屏幕预览与打印共用 [src/components/planner](src/components/planner) 下的同一套组件，不要为打印另写一份排版。
- 胶带生成器的绘制与导出逻辑放在 [src/utils/tape.ts](src/utils/tape.ts)，尺寸同样以 mm 为单位、按 DPI 换算像素；PSD 依赖 `ag-psd` 只在导出时动态 import，不要改成首屏静态引入。

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

### 修改代码后 dev 页面未实时更新

优先按以下顺序排查：

1. 先强制刷新页面（Ctrl+Shift+R）。
2. 确认仅有一个 dev 进程在运行且端口一致。
3. 重启 `npm run dev`。
4. 若仍偶发，使用轮询监听启动：

```bash
CHOKIDAR_USEPOLLING=true npm run dev -- --host 0.0.0.0 --port 5173
```

### 图表重复渲染或内存占用增长

检查图表组件是否在 onUnmounted 时调用 dispose。

## 发布前检查

1. npm run build 成功。
2. 核心路径手动验证：商品新增、订单新增、订单编辑、订单删除、定价规则保存。
3. 验证支出记录页的新增/编辑/删除流程。
4. 验证收入记录页的集市新增/编辑/删除、线上收入保存、总收入与利润计算，以及利润为负时的标红提醒。
5. 验证内页生成器的尺寸切换、模块增删与参数调整、模板保存/载入，以及打印预览中的纸张尺寸是否与所选 mm 一致。
6. 验证胶带生成页的宽长切换、图案层增删与排数/列数/大小调整、透明底，以及 PNG 与 PSD 导出的画布尺寸和图层结构。
7. 验证研究所与灵感页的新增/编辑/删除流程。
8. 确认浏览器刷新后本地数据仍可读。
