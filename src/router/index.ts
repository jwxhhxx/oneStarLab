import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
  history: createWebHistory(),
  routes: [
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
      path: '/data',
      name: 'data',
      component: () => import('@/views/DataView.vue'),
      meta: { title: '数据中心' },
    },
  ],
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title ?? 'OneStarLab')} · OneStarLab`;
});

export default router;