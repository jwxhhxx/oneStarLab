<script setup lang="ts">
import { Goods, Money, Tickets, WarningFilled } from '@element-plus/icons-vue';
import * as echarts from 'echarts';
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

import { useShopStore } from '@/stores/useShopStore';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const chartRef = ref<HTMLDivElement | null>(null);
let chart: echarts.ECharts | null = null;

const metrics = computed(() => [
  { label: '商品数', value: `${store.stats.totalProducts}`, tip: '当前在售商品', icon: Goods, accent: 'blue' },
  { label: '订单数', value: `${store.stats.totalOrders}`, tip: '累计已录入订单', icon: Tickets, accent: 'pink' },
  { label: '累计销售额', value: formatCurrency(store.stats.totalSales), tip: '订单实收总额', icon: Money, accent: 'gold' },
  { label: '低库存提醒', value: `${store.stats.lowStockCount}`, tip: '建议优先补货处理', icon: WarningFilled, accent: 'mint' },
]);

const insightList = computed(() => [
  { title: '本月销售额', tip: '本月累计成交', value: formatCurrency(store.stats.monthSales) },
  { title: '累计净利润', tip: '扣除商品成本和费用后', value: formatCurrency(store.stats.totalProfit) },
  { title: '库存预警', tip: '库存 ≤ 安全库存的商品数', value: `${store.stats.lowStockCount} 个` },
]);

function renderChart() {
  if (!chartRef.value) return;

  if (!chart) {
    chart = echarts.init(chartRef.value);
  }

  chart.setOption({
    color: ['#6c7cff', '#f08cab'],
    tooltip: { trigger: 'axis' },
    legend: { data: ['销售额', '净利润'] },
    grid: { left: 24, right: 18, top: 48, bottom: 24, containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: store.monthlyTrend.map((item) => item.month),
      axisLine: { lineStyle: { color: '#d7deee' } },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#eef2fb' } },
    },
    series: [
      {
        name: '销售额',
        type: 'line',
        smooth: true,
        areaStyle: { color: 'rgba(108, 124, 255, 0.12)' },
        data: store.monthlyTrend.map((item) => item.sales),
      },
      {
        name: '净利润',
        type: 'line',
        smooth: true,
        areaStyle: { color: 'rgba(240, 140, 171, 0.10)' },
        data: store.monthlyTrend.map((item) => item.profit),
      },
    ],
  });
}

function handleResize() {
  chart?.resize();
}

watch(
  () => store.monthlyTrend,
  async () => {
    await nextTick();
    renderChart();
  },
  { deep: true },
);

onMounted(async () => {
  await store.initialize();
  await nextTick();
  renderChart();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  chart?.dispose();
});
</script>

<template>
  <div class="page-stack">
    <section class="page-hero">
      <div>
        <div class="hero-kicker">OneStarLab · 今日总览</div>
        <h2 class="hero-title">先处理低库存，再看利润趋势</h2>
        <p class="hero-description">
          这版界面更偏轻柔手帐风，重点信息会更突出，打开后可以更快判断今天该补货还是该调价。
        </p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">本月销售 {{ formatCurrency(store.stats.monthSales) }}</span>
        <span class="hero-badge">预警 {{ store.stats.lowStockCount }} 项</span>
      </div>
    </section>

    <div class="metric-grid">
      <el-card v-for="item in metrics" :key="item.label" class="metric-card section-card">
        <div class="metric-top">
          <div class="metric-icon" :class="`metric-icon--${item.accent}`">
            <el-icon><component :is="item.icon" /></el-icon>
          </div>
          <el-tag size="small" round type="info">实时</el-tag>
        </div>
        <div class="metric-label">{{ item.label }}</div>
        <div class="metric-value">{{ item.value }}</div>
        <div class="inline-tip">{{ item.tip }}</div>
      </el-card>
    </div>

    <div class="grid-two">
      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>近 6 个月经营趋势</h3>
            <span class="inline-tip">用更柔和的趋势图查看销售额与利润变化</span>
          </div>
        </template>
        <div ref="chartRef" class="chart-box" />
      </el-card>

      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>经营提醒</h3>
            <span class="inline-tip">当天重点一眼可见</span>
          </div>
        </template>

        <div class="insight-list">
          <div v-for="item in insightList" :key="item.title" class="insight-item">
            <div>
              <strong>{{ item.title }}</strong>
              <div class="inline-tip">{{ item.tip }}</div>
            </div>
            <div class="insight-value">{{ item.value }}</div>
          </div>
        </div>
      </el-card>
    </div>

    <div class="grid-two">
      <el-card class="section-card soft-table">
        <template #header>
          <div class="toolbar">
            <h3>低库存商品</h3>
            <span class="inline-tip">库存 ≤ 安全库存的商品</span>
          </div>
        </template>

        <div class="table-scroll">
          <el-table :data="store.lowStockProducts" empty-text="暂无预警商品">
            <el-table-column prop="name" label="商品" min-width="180" />
            <el-table-column prop="sku" label="SKU" width="140" />
            <el-table-column label="当前库存" width="100">
              <template #default="{ row }">
                <el-tag type="danger" round>{{ row.stock }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="safeStock" label="安全库存" width="100" />
          </el-table>
        </div>
      </el-card>

      <el-card class="section-card soft-table">
        <template #header>
          <div class="toolbar">
            <h3>最近订单</h3>
            <span class="inline-tip">快速查看最近录入情况</span>
          </div>
        </template>

        <div class="table-scroll">
          <el-table :data="store.recentOrders" empty-text="暂无订单">
            <el-table-column prop="orderNo" label="订单号" min-width="160" />
            <el-table-column prop="channel" label="渠道" width="90" />
            <el-table-column prop="customerName" label="客户" width="110" />
            <el-table-column label="净利润" width="110">
              <template #default="{ row }">
                <span class="insight-value">{{ formatCurrency(row.netProfit) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-card>
    </div>
  </div>
</template>