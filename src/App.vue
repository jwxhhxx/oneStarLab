<script setup lang="ts">
import { Calendar, DataAnalysis, EditPen, Goods, MagicStick, Money, Tickets } from '@element-plus/icons-vue';
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

const menus = [
  { path: '/', label: '经营看板', icon: DataAnalysis },
  { path: '/products', label: '商品中心', icon: Goods },
  { path: '/orders', label: '订单中心', icon: Tickets },
  { path: '/pricing', label: '自动定价', icon: Money },
  { path: '/lab', label: '研究所', icon: EditPen },
  { path: '/inspiration', label: '灵感生成', icon: MagicStick },
];

const pageTitle = computed(() => String(route.meta.title ?? '店铺经营后台'));
const todayText = computed(() =>
  new Intl.DateTimeFormat('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(new Date()),
);
</script>

<template>
  <el-container class="app-shell">
    <el-aside width="240px" class="app-sidebar">
      <div class="brand-panel">
        <div class="brand-chip">OneStarLab</div>
        <div class="brand-title">手帐店铺助手</div>
        <div class="brand-subtitle">库存、利润、定价都放进一个清爽后台。</div>
      </div>

      <el-menu router :default-active="route.path" class="side-menu">
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-footer">
        <div class="sidebar-pill">本地离线模式</div>
        <p>数据仅保存在当前浏览器，建议定期导出备份。</p>
      </div>
    </el-aside>

    <el-container class="app-content-shell">
      <el-header class="app-header">
        <div class="page-intro">
          <div class="header-kicker">
            <el-icon><Calendar /></el-icon>
            <span>{{ todayText }}</span>
          </div>
          <h1>{{ pageTitle }}</h1>
          <p>轻量、柔和、适合个人店主每天打开就能直接工作的经营界面。</p>
        </div>

        <div class="header-actions">
          <el-tag round effect="dark" type="primary">纯前端运行</el-tag>
          <el-tag round type="success">IndexedDB 已启用</el-tag>
        </div>
      </el-header>

      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>