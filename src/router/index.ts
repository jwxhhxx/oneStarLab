import { createRouter, createWebHashHistory } from 'vue-router';

import { useAuthStore } from '@/stores/useAuthStore';

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { title: '登录', public: true },
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { title: '经营看板' },
    },
    {
      path: '/products',
      name: 'products',
      component: () => import('@/views/ProductsView.vue'),
      meta: { title: '商品中心' },
    },
    {
      path: '/orders',
      name: 'orders',
      component: () => import('@/views/OrdersView.vue'),
      meta: { title: '订单中心' },
    },
    {
      path: '/income',
      name: 'income',
      component: () => import('@/views/IncomeView.vue'),
      meta: { title: '收入记录' },
    },
    {
      path: '/markets',
      redirect: '/income',
    },
    {
      path: '/pricing',
      name: 'pricing',
      component: () => import('@/views/PricingView.vue'),
      meta: { title: '自动定价' },
    },
    {
      path: '/lab',
      name: 'lab',
      component: () => import('@/views/LabView.vue'),
      meta: { title: '研究所' },
    },
    {
      path: '/inspiration',
      name: 'inspiration',
      component: () => import('@/views/InspirationView.vue'),
      meta: { title: '灵感生成' },
    },
    {
      path: '/scan',
      name: 'scan',
      component: () => import('@/views/BarcodeScanView.vue'),
      meta: { title: '扫码出入库' },
    },
    {
      path: '/data',
      name: 'data',
      component: () => import('@/views/DataView.vue'),
      meta: { title: '数据中心' },
    },
    {
      path: '/expenses',
      name: 'expenses',
      component: () => import('@/views/ExpensesView.vue'),
      meta: { title: '支出记录' },
    },
    {
      path: '/halftone',
      name: 'halftone',
      component: () => import('@/views/HalftoneView.vue'),
      meta: { title: '印章生成' },
    },
    {
      path: '/planner',
      name: 'planner',
      component: () => import('@/views/PlannerView.vue'),
      meta: { title: '内页生成器' },
    },
    {
      path: '/tape',
      name: 'tape',
      component: () => import('@/views/TapeView.vue'),
      meta: { title: '胶带生成' },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { title: '设置' },
    },
  ],
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.checked) {
    await auth.check();
  }
  if (to.meta.public) {
    if (to.name === 'login' && auth.authenticated) {
      return { path: '/' };
    }
    return true;
  }
  if (!auth.authenticated) {
    return {
      name: 'login',
      query: to.fullPath === '/' ? {} : { redirect: to.fullPath },
    };
  }
  return true;
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title ?? 'OneStarLab')} · OneStarLab`;
});

export default router;