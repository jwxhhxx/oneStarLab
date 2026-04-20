# OneStarLab 手帐店铺助手

OneStarLab 是一个纯前端实现的手帐店铺经营后台（SPA），数据保存在浏览器 IndexedDB。

在线访问地址。<https://jwxhhxx.github.io/oneStarLab/>

## 功能概览

当前版本支持：

- 商品与 SKU 管理
- 库存追踪与低库存预警
- 订单录入与利润核算
- 开店支出记录与累计统计
- 自动定价建议
- 基础经营看板
- 研发流程管理（研究所）
- 灵感随机生成功能

## 页面说明

- 经营看板：查看销售额、利润、订单趋势等核心指标
- 商品中心：维护商品、库存、安全库存与建议售价
- 订单中心：录入与维护订单，自动计算成本和净利润
- 支出记录：记录开店以来的支出用途与金额，形成成本台账
- 自动定价：维护目标毛利率、平台费率和取整策略
- 研究所：管理灵感到上新的项目进度
- 灵感生成：按品类抽取命名、风格、主题等灵感

## 本地启动

```bash
npm install
npm run dev
```

默认访问地址：

- <http://localhost:5173>
- GitHub Pages：<https://jwxhhxx.github.io/oneStarLab/>

## 构建与预览

```bash
npm run build
npm run preview
```

说明：

- build 命令会先执行 TypeScript 类型检查（vue-tsc），再进行 Vite 打包。
- 当前项目未配置独立 lint/test 脚本。

## 技术栈

- Vue 3
- TypeScript
- Vite
- Pinia
- Element Plus
- Dexie / IndexedDB
- ECharts

## 项目结构

```text
src/
  db/         IndexedDB 数据库定义（Dexie）
  router/     路由与页面标题策略
  stores/     Pinia 业务编排层
  types/      业务类型定义
  utils/      纯函数工具（如定价计算）
  views/      页面视图层
```

## 数据与持久化说明

- 本项目无后端服务，所有数据仅保存在当前浏览器本地。
- 清除浏览器站点数据、切换设备/浏览器后，数据不会自动同步。
- 建议在生产使用前准备导入导出或备份机制。

## 相关文档

- 架构与边界说明：[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- 开发与维护指南：[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
