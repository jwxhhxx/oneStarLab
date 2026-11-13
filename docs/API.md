# OneStarLab REST API

所有接口以 `/api/v1` 为前缀，请求与响应均为 JSON。字段使用 camelCase，与前端 `src/types/index.ts` 保持一致。

## 认证与通用约定

- 默认无认证；设置 `API_TOKEN` 环境变量后，所有 `/api/v1` 接口（包括健康检查）需携带 `Authorization: Bearer <token>`。
- 成功：`200` 返回资源对象、数组或 `{ "ok": true }`；删除成功返回 `{ "ok": true }`。
- 失败：非 2xx，响应体为 `{ "error": "中文错误信息" }`。
- 请求体上限 64MB，未知字段会被拒绝（`400`）。
- 时间均为 UTC ISO 8601 字符串（毫秒精度），例如 `2026-11-13T16:48:09.123Z`。

错误码：

| 状态码 | 场景 |
| --- | --- |
| `400` | 参数/业务校验失败（如库存不足、默认分类不可删除、金额为负） |
| `404` | 资源不存在 |
| `500` | 服务端错误（数据库异常等） |

## 健康检查与初始化

### GET /health

```json
{ "status": "ok", "time": "2026-11-13T16:48:09Z" }
```

### GET /bootstrap

一次返回页面所需的全部业务数据：

```json
{
  "products": [],
  "categories": [],
  "orders": [],
  "pricingRule": { "id": 1, "targetMargin": 0.45, "channelFeeRate": 0.05, "extraCost": 1.2, "roundingMode": "x.9" },
  "labInspirations": [],
  "labProjects": [],
  "drawingInspirations": [],
  "expenses": [],
  "inventoryTransactions": [],
  "markets": [],
  "onlineIncome": 0,
  "onlineIncomeUpdatedAt": "",
  "plannerTemplates": []
}
```

## 商品与分类

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/products` | 商品列表 |
| POST | `/products` | 新建商品，返回商品；`sku` 为空时服务端生成，`salePrice<=0` 时用建议价 |
| PUT | `/products/{id}` | 编辑商品 |
| DELETE | `/products/{id}` | 删除商品 |
| GET | `/products/sku/{sku}` | 按 SKU/条码查找，未找到返回 404 |
| POST | `/sku` | 预生成 SKU（不创建商品） |
| GET | `/categories` | 分类列表 |
| POST | `/categories` | 新建分类，body `{ "name": "胶带" }` |
| DELETE | `/categories/{id}?reassignTo=胶带` | 删除分类；默认分类「未分类」不可删；`reassignTo` 为空时该分类商品归入未分类 |

商品请求体：

```json
{
  "name": "和纸胶带",
  "sku": "",
  "category": "胶带",
  "supplier": "供应商",
  "purchaseCost": 3,
  "packagingCost": 0.5,
  "salePrice": 9.9,
  "stock": 12,
  "safeStock": 3,
  "note": ""
}
```

SKU 预生成：

```json
{ "category": "胶带", "excludeProductId": 0, "createIfMissing": true }
```

返回 `{ "sku": "1-0001" }`。

## 订单

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/orders` | 订单列表（按创建时间倒序） |
| POST | `/orders` | 新建订单，自动生成单号并扣库存 |
| PUT | `/orders/{id}` | 编辑订单（回补旧库存，沿用只处理第一件商品的既有语义） |
| DELETE | `/orders/{id}` | 删除订单并回补库存 |

请求体：

```json
{
  "productId": 1,
  "customerName": "顾客",
  "channel": "微信",
  "quantity": 2,
  "discountAmount": 0,
  "shippingCost": 0,
  "platformFeeRate": 0.05
}
```

服务端计算：`totalAmount = max(0, salePrice * quantity - discountAmount)`，`platformFee = totalAmount * platformFeeRate`，`goodsCost = (purchaseCost + packagingCost) * quantity`，`netProfit = totalAmount - shippingCost - platformFee - goodsCost`。

## 定价规则

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/pricing-rule` | 读取规则（缺省时返回默认定价规则） |
| PUT | `/pricing-rule` | 保存规则 |

```json
{ "targetMargin": 0.45, "channelFeeRate": 0.05, "extraCost": 1.2, "roundingMode": "x.9" }
```

`roundingMode` 支持 `"whole"`（向上取整）与 `"x.9"`（尾数 .9）。

## 库存与扫码

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/inventory/transactions` | 流水列表 |
| POST | `/inventory/transactions` | 直接新增一条流水并调整库存 |
| DELETE | `/inventory/transactions/{id}` | 删除流水并回滚库存 |
| POST | `/inventory/process` | 单条入/出库，出库时自动建单；返回 `{ "ok": true }` |
| POST | `/inventory/scan` | 批量扫码，返回 `{ "orderNo": "ORD-..." }` |

`/inventory/process` 请求体：

```json
{ "productId": 1, "type": "out", "quantity": 1, "barcode": "1-0001", "note": "" }
```

`/inventory/scan` 请求体：

```json
{
  "type": "out",
  "note": "",
  "items": [{ "productId": 1, "quantity": 2, "barcode": "1-0001", "note": "" }]
}
```

`type` 为 `in`（入库）或 `out`（出库）。出库库存不足返回 `400`，错误信息形如 `商品 和纸胶带 库存不足`。

## 研究所与灵感

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET / POST | `/lab/inspirations` | 灵感列表 / 新建 |
| PUT / DELETE | `/lab/inspirations/{id}` | 编辑 / 删除 |
| GET / POST | `/lab/projects` | 研发项目列表 / 新建 |
| PUT / DELETE | `/lab/projects/{id}` | 编辑 / 删除 |
| PATCH | `/lab/projects/{id}/stage` | 修改阶段，body `{ "stage": "打样中" }` |
| POST | `/lab/projects/{id}/proofs` | 追加打样记录 |
| POST | `/lab/projects/{id}/convert` | 转为正式商品，SKU 为 `LAB-<id>` |
| GET / POST | `/drawings` | 绘画灵感列表 / 新建 |
| PUT / DELETE | `/drawings/{id}` | 编辑 / 删除 |

阶段与进度：灵感整理 10、设计中 35、打样中 60、排产中 82、待上新 95、已上新 100；进度取已到达阶段的最大值。

打样记录请求体：

```json
{ "title": "第一版打样", "date": "2026-11-13", "note": "", "image": "" }
```

## 支出、集市与线上收入

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET / POST | `/expenses` | 支出列表 / 新建 `{ "purpose": "采购", "amount": 120 }` |
| PUT / DELETE | `/expenses/{id}` | 编辑 / 删除 |
| GET / POST | `/markets` | 集市列表 / 新建 |
| PUT / DELETE | `/markets/{id}` | 编辑 / 删除 |
| GET | `/settings/online-income` | 返回 `{ "onlineIncome": 800, "updatedAt": "..." }` |
| PUT | `/settings/online-income` | body `{ "amount": 800 }`，负数返回 400 |

集市请求体：

```json
{ "name": "跨年手帐市集", "date": "2026-01-01", "location": "文创园", "income": 500, "note": "" }
```

## 内页模板

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/planner-templates` | 模板列表（按更新时间倒序） |
| POST | `/planner-templates` | 新建模板 |
| PUT | `/planner-templates/{id}` | 编辑模板 |
| DELETE | `/planner-templates/{id}` | 删除模板 |

请求体与 `PlannerTemplateInput` 一致：`name`、`sizeKey`、`widthMm`、`heightMm`、`paddingMm`、`accent`、`paper`、`modules`。

## 备份与恢复

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/export` | 下载 `onestarlab-backup.json`（`Content-Disposition: attachment`） |
| POST | `/import` | 覆盖导入快照，body 为导出的 JSON 全文 |

快照结构：

```json
{
  "version": 1,
  "exportedAt": "2026-11-13T16:48:09.123Z",
  "products": [],
  "categories": [],
  "orders": [],
  "pricingRule": { "...": "..." },
  "labInspirations": [],
  "labProjects": [],
  "drawingInspirations": [],
  "expenses": [],
  "inventoryTransactions": [],
  "markets": [],
  "settings": [],
  "plannerTemplates": []
}
```

导入会清空现有业务表后按原 id 恢复，整个过程在一个事务中，失败自动回滚。成功返回 `{"ok": true, "source": "snapshot"}`。

### 旧版浏览器备份

`POST /import` 也接受旧版纯前端版本（IndexedDB / Dexie）在数据中心导出的备份文件：

```json
{
  "formatName": "dexie",
  "formatVersion": 1,
  "data": {
    "databaseName": "oneStarLabDb",
    "databaseVersion": 9,
    "tables": [{ "name": "products", "schema": "++id,sku", "rowCount": 2 }],
    "data": [{ "tableName": "products", "inbound": true, "rows": [{ "id": 1, "name": "…" }] }]
  }
}
```

服务端按表名映射到新版结构（商品、分类、订单、定价规则、研究所、支出、库存流水、集市、线上收入、内页模板），忽略 `$types` 等 Dexie 元数据；导入成功返回 `{"ok": true, "source": "dexie"}`。

- 文件无法识别时返回 `400`，例如 `{"error":"备份中没有可识别的 OneStarLab 数据表"}`。
- 旧版备份中缺少的表按空表处理，不会报错。
- 导入同样是覆盖式，建议先导出当前快照。

## curl 示例

```bash
curl http://127.0.0.1:8080/api/v1/health
curl http://127.0.0.1:8080/api/v1/bootstrap
curl -X POST http://127.0.0.1:8080/api/v1/expenses \
  -H "Content-Type: application/json" \
  -d '{"purpose":"采购","amount":120}'
curl -OJ http://127.0.0.1:8080/api/v1/export
```

启用令牌时追加：

```bash
-H "Authorization: Bearer $API_TOKEN"
```
