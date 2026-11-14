<script setup lang="ts">
import { Brush, Calendar, Coin, DataAnalysis, EditPen, Files, Goods, MagicStick, Memo, Menu, Money, Picture, Setting, SwitchButton, Tickets } from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useAuthStore } from '@/stores/useAuthStore';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const isMobile = ref(false);
const mobileMenuVisible = ref(false);

const isAuthPage = computed(() => route.name === 'login');

async function handleLogout() {
  try {
    await ElMessageBox.confirm('退出后需要重新输入口令才能进入。', '退出登录', {
      confirmButtonText: '退出登录',
      cancelButtonText: '取消',
      type: 'warning',
    });
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    return;
  }
  await authStore.logout();
  mobileMenuVisible.value = false;
  await router.replace({ name: 'login' });
}

let mediaQuery: MediaQueryList | null = null;

function handleViewportChange(event: MediaQueryListEvent) {
  isMobile.value = event.matches;
  if (!event.matches) {
    mobileMenuVisible.value = false;
  }
}

onMounted(() => {
  mediaQuery = window.matchMedia('(max-width: 768px)');
  isMobile.value = mediaQuery.matches;
  mediaQuery.addEventListener('change', handleViewportChange);
});

onUnmounted(() => {
  mediaQuery?.removeEventListener('change', handleViewportChange);
});

watch(
  () => route.path,
  () => {
    mobileMenuVisible.value = false;
  },
);

const menus = [
  { path: '/', label: '经营看板', icon: DataAnalysis },
  { path: '/products', label: '商品中心', icon: Goods },
  { path: '/orders', label: '订单中心', icon: Tickets },
  { path: '/income', label: '收入记录', icon: Coin },
  { path: '/pricing', label: '自动定价', icon: Money },
  { path: '/expenses', label: '支出记录', icon: Money },
  { path: '/scan', label: '扫码出入库', icon: Tickets },
  { path: '/lab', label: '研究所', icon: EditPen },
  { path: '/inspiration', label: '灵感生成', icon: MagicStick },
  { path: '/halftone', label: '印章生成', icon: Picture },
  { path: '/planner', label: '内页生成器', icon: Memo },
  { path: '/tape', label: '胶带生成', icon: Brush },
  { path: '/data', label: '数据中心', icon: Files },
  { path: '/settings', label: '设置', icon: Setting },
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
  <router-view v-if="isAuthPage" />
  <el-container v-else class="app-shell">
    <el-aside v-if="!isMobile" width="240px" class="app-sidebar">
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
        <div class="sidebar-pill">服务端存储</div>
        <p>数据保存在服务端 SQLite，建议定期导出备份。</p>
        <el-button class="sidebar-logout" text type="danger" :icon="SwitchButton" @click="handleLogout">
          退出登录
        </el-button>
      </div>
    </el-aside>

    <el-container class="app-content-shell">
      <div v-if="isMobile" class="mobile-topbar">
        <div class="mobile-brand">
          <div class="brand-chip">OneStarLab</div>
          <strong>手帐店铺助手</strong>
        </div>
        <el-button circle @click="mobileMenuVisible = true">
          <el-icon><Menu /></el-icon>
        </el-button>
      </div>

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
          <el-tag round effect="dark" type="primary">服务端运行</el-tag>
          <el-tag round type="success">SQLite 已连接</el-tag>
        </div>
      </el-header>

      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>

    <el-drawer v-model="mobileMenuVisible" direction="ltr" size="82%" class="mobile-menu-drawer">
      <template #header>
        <div class="brand-panel" style="margin: 0">
          <div class="brand-chip">OneStarLab</div>
          <div class="brand-title">手帐店铺助手</div>
          <div class="brand-subtitle">库存、利润、定价都放进一个清爽后台。</div>
        </div>
      </template>

      <el-menu router :default-active="route.path" class="side-menu mobile-side-menu">
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>

      <div class="mobile-drawer-footer">
        <el-button text type="danger" :icon="SwitchButton" @click="handleLogout">退出登录</el-button>
      </div>
    </el-drawer>
  </el-container>
</template>