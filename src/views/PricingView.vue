<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';

import { useShopStore } from '@/stores/useShopStore';
import type { PricingRule } from '@/types';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const selectedProductId = ref<number>(0);

const ruleForm = reactive<PricingRule>({
  id: 1,
  targetMargin: 0.45,
  channelFeeRate: 0.05,
  extraCost: 1.2,
  roundingMode: 'x.9',
});

watch(
  () => store.pricingRule,
  (value) => {
    Object.assign(ruleForm, value);
  },
  { immediate: true, deep: true },
);

const selectedProduct = computed(() => store.products.find((item) => item.id === selectedProductId.value));

const preview = computed(() => {
  if (!selectedProduct.value) return null;

  return {
    currentPrice: selectedProduct.value.salePrice,
    suggestedPrice: store.getSuggestedPrice(selectedProduct.value),
    minPrice: store.getMinPrice(selectedProduct.value),
  };
});

const summaryCards = computed(() => [
  { label: '目标毛利率', value: `${Math.round(ruleForm.targetMargin * 100)}%` },
  { label: '渠道费率', value: `${Math.round(ruleForm.channelFeeRate * 100)}%` },
  { label: '固定附加成本', value: formatCurrency(ruleForm.extraCost) },
  { label: '尾数规则', value: ruleForm.roundingMode === 'x.9' ? '尾数 .9' : ruleForm.roundingMode === 'whole' ? '向上取整' : '不处理' },
]);

async function saveRule() {
  await store.savePricingRule({ ...ruleForm });
  ElMessage.success('定价规则已更新');
}

onMounted(() => {
  store.initialize();
});
</script>

<template>
  <div class="page-stack">
    <section class="page-hero">
      <div>
        <div class="hero-kicker">自动定价 · 更柔和的价格面板</div>
        <h2 class="hero-title">用同一套规则控制利润底线与建议售价</h2>
        <p class="hero-description">适合在调价前先试算，避免凭感觉定价，保证每个渠道都有利润空间。</p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">毛利率 {{ Math.round(ruleForm.targetMargin * 100) }}%</span>
        <span class="hero-badge">费率 {{ Math.round(ruleForm.channelFeeRate * 100) }}%</span>
      </div>
    </section>

    <div class="summary-grid">
      <div v-for="item in summaryCards" :key="item.label" class="summary-tile">
        <strong>{{ item.label }}</strong>
        <span>{{ item.value }}</span>
      </div>
    </div>

    <div class="grid-two">
      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>定价规则</h3>
            <span class="inline-tip">系统会按规则计算建议售价</span>
          </div>
        </template>

        <el-form label-width="120px">
          <el-form-item label="目标毛利率">
            <el-input-number v-model="ruleForm.targetMargin" :min="0.1" :max="0.9" :step="0.01" />
          </el-form-item>
          <el-form-item label="渠道费率">
            <el-input-number v-model="ruleForm.channelFeeRate" :min="0" :max="0.5" :step="0.01" />
          </el-form-item>
          <el-form-item label="固定附加成本">
            <el-input-number v-model="ruleForm.extraCost" :min="0" :step="0.1" />
          </el-form-item>
          <el-form-item label="尾数规则">
            <el-select v-model="ruleForm.roundingMode">
              <el-option label="不处理" value="none" />
              <el-option label="尾数 .9" value="x.9" />
              <el-option label="向上取整" value="whole" />
            </el-select>
          </el-form-item>
          <el-button type="primary" @click="saveRule">保存规则</el-button>
        </el-form>
      </el-card>

      <el-card class="section-card">
        <template #header>
          <div class="toolbar">
            <h3>价格模拟器</h3>
            <span class="inline-tip">选一个商品试算建议售价</span>
          </div>
        </template>

        <el-form label-width="90px">
          <el-form-item label="选择商品">
            <el-select v-model="selectedProductId" placeholder="请选择商品">
              <el-option v-for="item in store.products" :key="item.id" :label="item.name" :value="item.id!" />
            </el-select>
          </el-form-item>
        </el-form>

        <div v-if="preview" class="result-box">
          <div class="result-item">
            <strong>当前售价</strong>
            <span>{{ formatCurrency(preview.currentPrice) }}</span>
          </div>
          <div class="result-item">
            <strong>建议售价</strong>
            <span class="insight-value">{{ formatCurrency(preview.suggestedPrice) }}</span>
          </div>
          <div class="result-item">
            <strong>最低售价</strong>
            <span>{{ formatCurrency(preview.minPrice) }}</span>
          </div>
        </div>
        <el-empty v-else description="选择商品后查看定价建议" />
      </el-card>
    </div>

    <el-card class="section-card soft-table">
      <template #header>
        <div class="toolbar">
          <h3>商品价格建议表</h3>
          <span class="inline-tip">可作为后续调价参考</span>
        </div>
      </template>

      <div class="table-scroll">
        <el-table :data="store.products">
          <el-table-column prop="name" label="商品" min-width="180" />
          <el-table-column prop="sku" label="SKU" width="140" />
          <el-table-column label="当前售价" width="120">
            <template #default="{ row }">{{ formatCurrency(row.salePrice) }}</template>
          </el-table-column>
          <el-table-column label="建议售价" width="120">
            <template #default="{ row }">
              <span class="insight-value">{{ formatCurrency(store.getSuggestedPrice(row)) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="最低售价" width="120">
            <template #default="{ row }">{{ formatCurrency(store.getMinPrice(row)) }}</template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>