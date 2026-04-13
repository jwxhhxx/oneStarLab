<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';

import { useShopStore } from '@/stores/useShopStore';
import type { NewOrderInput } from '@/types';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const dialogVisible = ref(false);

const emptyForm = (): NewOrderInput => ({
  productId: 0,
  customerName: '',
  channel: '微信',
  quantity: 1,
  discountAmount: 0,
  shippingCost: 0,
  platformFeeRate: 0.05,
});

const form = reactive<NewOrderInput>(emptyForm());

const selectedProduct = computed(() => store.products.find((item) => item.id === form.productId));
const preview = computed(() => {
  if (!selectedProduct.value) {
    return { total: 0, cost: 0, fee: 0, profit: 0 };
  }

  const total = Math.max(0, selectedProduct.value.salePrice * form.quantity - form.discountAmount);
  const cost = (selectedProduct.value.purchaseCost + selectedProduct.value.packagingCost) * form.quantity;
  const fee = total * form.platformFeeRate;
  const profit = total - cost - fee - form.shippingCost;

  return {
    total,
    cost,
    fee,
    profit,
  };
});

const summaryCards = computed(() => {
  const avgOrder = store.orders.length
    ? store.orders.reduce((sum, item) => sum + item.totalAmount, 0) / store.orders.length
    : 0;

  return [
    { label: '累计订单数', value: `${store.orders.length}` },
    { label: '累计销售额', value: formatCurrency(store.stats.totalSales) },
    { label: '累计净利润', value: formatCurrency(store.stats.totalProfit) },
    { label: '平均客单价', value: formatCurrency(avgOrder) },
  ];
});

function resetForm() {
  Object.assign(form, emptyForm());
}

async function submitOrder() {
  try {
    await store.addOrder({ ...form });
    ElMessage.success('订单已保存并自动扣减库存');
    dialogVisible.value = false;
    resetForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '订单保存失败');
  }
}

onMounted(() => {
  store.initialize();
});
</script>

<template>
  <div class="page-stack">
    <section class="page-hero page-hero--violet">
      <div>
        <div class="hero-kicker">订单中心 · 更清楚的利润视角</div>
        <h2 class="hero-title">录单后立刻看到销售额、成本和净利润</h2>
        <p class="hero-description">这版更强调订单结果反馈，方便你快速判断哪一单更赚钱。</p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">订单 {{ store.orders.length }} 笔</span>
        <span class="hero-badge">总利润 {{ formatCurrency(store.stats.totalProfit) }}</span>
      </div>
    </section>

    <div class="summary-grid">
      <div v-for="item in summaryCards" :key="item.label" class="summary-tile">
        <strong>{{ item.label }}</strong>
        <span>{{ item.value }}</span>
      </div>
    </div>

    <el-card class="section-card soft-table">
      <div class="toolbar">
        <div>
          <h2>订单中心</h2>
          <div class="inline-tip">录单后自动计算利润并扣减库存</div>
        </div>
        <el-button type="primary" @click="dialogVisible = true">新建订单</el-button>
      </div>

      <el-table :data="store.orders" empty-text="还没有订单，先录入一笔。">
        <el-table-column prop="orderNo" label="订单号" min-width="170" />
        <el-table-column prop="channel" label="渠道" width="100" />
        <el-table-column prop="customerName" label="客户" width="120" />
        <el-table-column label="商品" min-width="180">
          <template #default="{ row }">{{ row.items[0]?.productName ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="销售额" width="110">
          <template #default="{ row }">{{ formatCurrency(row.totalAmount) }}</template>
        </el-table-column>
        <el-table-column label="商品成本" width="110">
          <template #default="{ row }">{{ formatCurrency(row.goodsCost) }}</template>
        </el-table-column>
        <el-table-column label="净利润" width="110">
          <template #default="{ row }">
            <span class="insight-value">{{ formatCurrency(row.netProfit) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="录入订单" width="720px">
      <el-form label-width="92px">
        <div class="form-grid">
          <el-form-item label="选择商品">
            <el-select v-model="form.productId" placeholder="请选择商品">
              <el-option
                v-for="item in store.products"
                :key="item.id"
                :label="`${item.name}（库存 ${item.stock}）`"
                :value="item.id!"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="客户昵称"><el-input v-model="form.customerName" /></el-form-item>
          <el-form-item label="销售渠道">
            <el-select v-model="form.channel">
              <el-option label="微信" value="微信" />
              <el-option label="小红书" value="小红书" />
              <el-option label="淘宝" value="淘宝" />
              <el-option label="线下" value="线下" />
            </el-select>
          </el-form-item>
          <el-form-item label="数量"><el-input-number v-model="form.quantity" :min="1" /></el-form-item>
          <el-form-item label="优惠金额"><el-input-number v-model="form.discountAmount" :min="0" :step="0.1" /></el-form-item>
          <el-form-item label="运费支出"><el-input-number v-model="form.shippingCost" :min="0" :step="0.1" /></el-form-item>
          <el-form-item label="平台费率"><el-input-number v-model="form.platformFeeRate" :min="0" :max="1" :step="0.01" /></el-form-item>
        </div>
      </el-form>

      <div class="result-box">
        <div class="result-item">
          <strong>预计销售额</strong>
          <span>{{ formatCurrency(preview.total) }}</span>
        </div>
        <div class="result-item">
          <strong>预计成本+费用</strong>
          <span>{{ formatCurrency(preview.cost + preview.fee + form.shippingCost) }}</span>
        </div>
        <div class="result-item">
          <strong>预计净利润</strong>
          <span class="insight-value">{{ formatCurrency(preview.profit) }}</span>
        </div>
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitOrder">保存订单</el-button>
      </template>
    </el-dialog>
  </div>
</template>