# OneStarLab — 业务说明与操作流程

本文件面向店主与运营人员，说明各个功能模块的业务意图与典型操作流程，帮助你理解系统如何在前端（浏览器）中完成日常经营任务。

## 概述

OneStarLab 是一个纯前端的经营后台，所有数据保存在浏览器的 IndexedDB（通过 Dexie 封装）。用户通过页面交互调用 Pinia store 的行为 (actions)，由 store 负责与 Dexie 读写并在必要时进行事务处理以保证数据一致性。

核心设计原则：
- 视图（views）不直接操作数据库，所有持久化逻辑封装在 `src/stores/useShopStore.ts`。
- 写入后统一调用 `loadAll()` 刷新 store 状态，确保 UI 可感知最新数据。
- 多表联动操作（例如出库同时写订单并更新库存）使用 Dexie 事务，保证原子性。

## 主要数据表（Dexie）

- `products`：商品信息（SKU、成本、售价、库存等）。
- `orders`：订单记录，含商品项、金额、净利润等。
- `pricingRules`：自动定价规则（目标毛利、平台费率、取整策略）。
- `labInspirations` / `labProjects`：研发/上新流程数据。
- `drawingInspirations`：灵感生成结果记录。
- `expenses`：开店支出记录。
- `markets`：集市出摊记录（名称、日期、地点、每场收入）。
- `settings`：键值配置表，当前保存线上收入（`onlineIncome`）与最后更新时间。
- `inventoryTransactions`：入/出库事务日志（扫码/手动/订单触发）。

## 页面与业务流说明

以下按页面列举核心业务行为与背后发生的逻辑：

### 经营看板（Dashboard）
- 展示由 `orders`、`products` 与 `expenses` 聚合得到的关键指标（总销售、净利润、月度趋势、低库存统计）。
- 数据由 store 的计算属性（`stats`、`monthlyTrend`）动态计算，不持久化。

### 商品中心（Products）
- 新增/修改商品会调用 `addProduct` / `updateProduct`：计算建议售价（`utils/pricing.ts`），写入 `products` 并刷新状态。
- 删除商品会从 `products` 删除记录并刷新。

### 订单中心（Orders）
- 创建订单时通过 `buildOrderPayload` 计算订单字段（orderNo、platformFee、goodsCost、netProfit 等）。
- 创建/更新/删除订单会在事务中调整 `products.stock`（保证库存与订单一致）。

### 自动定价（Pricing）
- 定价规则保存在 `pricingRules`（id=1），用于 `getSuggestedPrice` 计算建议价、`getMinPrice` 计算成本下限。

### 研究所与灵感（Lab / Inspiration）
- 研发项目与灵感生成分别持久化到 `labProjects`、`labInspirations` 与 `drawingInspirations`，便于跟踪进度与历史记录。

### 支出记录（Expenses）
- 向 `expenses` 写入时会记录 `purpose`、`amount`、`createdAt`，并在看板上汇总为成本统计。

### 收入记录（Income）
- 页面路径 `/income`（旧 `/markets` 会重定向过来），把线下集市收入与线上收入放在一起看。
- 集市部分：每场集市写入 `markets`，记录 `name`、`date`（YYYY-MM-DD）、`location`、`income`，可选 `note` 与系统写入的 `createdAt`；按日期倒序展示，支持按集市名称或地点搜索。
- 线上收入：不逐单记录，由店主在页面顶部输入框里按平台实际到账金额手工填写，点击保存后写入 `settings` 的 `onlineIncome` 键，并记录 `updatedAt` 作为“上次更新”时间；建议每周或每月更新一次。
- 汇总口径：总收入 = 集市累计收入 + 线上收入；总支出 = `expenses` 中所有 `amount` 之和；利润 = 总收入 − 总支出。
- 展示规则：页面顶部汇总卡片显示总收入、线上收入、总支出与利润；利润为负时利润数字、顶部徽章变红，并出现一条红色提醒，提示支出比收入多多少、需要补录或检查支出。
- 集市收入与手工线上收入都只用于收入页的台账与利润展示，不并入 `orders` 的销售额与净利润，避免和线上订单重复计算。

### 数据中心（导出/导入）
- 使用 `dexie-export-import` 的 `exportDB` / `importDB` 实现全库导出与导入。
- 导入会清除并重建表，请先在本地导出备份文件再进行导入操作；导出文件名默认含日期时间便于区分（示例：`onestarlab-backup-YYYY-MM-DD-HHmm.json`）。

### 条形码生成与扫码（Barcode / Scan）
- 商品中心内置的条码弹窗：使用 `JsBarcode` 生成 SVG 条码，支持将 SVG 转为 PNG 并通过 Web Share 或在新窗口打开以便手机保存；提供标签打印的简易 HTML 模板（打印时需根据标签机纸张调整尺寸）。
- `BarcodeScanView`：使用 `@zxing/library` 打开相机（优先后置摄像头），扫码结果可自动加入扫描队列（去抖与同 SKU 合并）。
- 批量扫码后点击“生成订单”会调用 store 的 `createOrderFromScans`：在单个 Dexie 事务内完成订单创建、库存更新与写入 `inventoryTransactions`，保证原子性。
- 单次扫码也可以调用 `processInventoryTransaction` 执行单条入/出库；若为出库，系统会在事务中自动生成对应订单（单笔）。

## 事务与一致性策略

- 对于会同时修改多张表的操作（如：生成订单并扣减库存，批量出库并写入交易日志），使用 Dexie 的 `db.transaction('rw', ...)` 将相关表列入事务范围，避免部分写入导致数据不一致。
- 写完成后统一调用 `loadAll()`，store 会从数据库重新拉取最新数据并更新响应式状态。

## 移动端与摄像头注意事项

- 摄像头扫码需要 HTTPS 或 `localhost`，且须在浏览器中允许相机权限。iOS Safari 在某些情况下需要用户手动切换到“后置摄像头”。
- PNG 下载：在支持 Web Share 的移动浏览器上会优先使用分享 API，作为回退会在新窗口打开图片以供用户长按保存。

## 常见业务示例

- 场景：收到拣货完成的多件商品，扫码逐个加入队列 → 点击“生成订单” → 系统在单次事务中创建订单（包含多条商品项）、减少对应库存并生成入/出库日志。
- 场景：门店临时补货，使用扫码选择入库模式（in）→ 批量扫码 → 点击生成订单（或直接执行入库），系统只会变更库存并写入入库交易（如需记录为采购订单，可在后续功能中扩展）。

## 限制与扩展建议

- 当前为单机离线应用，不支持多设备同步。如需多端协同，需要引入后端或同步层。
- 标签打印的最佳体验通常需要厂商的手机 App 或蓝牙打印机 SDK，Web 端的“打印”仅适用于支持打印的浏览器/平台；如需直接控制标签机，建议集成厂商 SDK 或通过中间服务桥接。
- 条码识别依赖摄像头质量与对焦，建议使用标准条码（CODE128 / EAN）并在条码生成时保证足够高度与对比度。

## 联系与部署

- 代码仓库：本项目位于 `./`（仓库根目录），部署到 GitHub Pages 可直接使用 GitHub Actions 自动构建与发布。
- 如需对接第三方系统（ERP、物流、打印机），建议先在 store 中定义外部适配接口，再实现具体同步或网关。

---

更多技术细节请参见：`src/stores/useShopStore.ts`（业务逻辑）、`src/db/appDb.ts`（Dexie schema）、`src/views`（页面实现）。
