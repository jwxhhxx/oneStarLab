<script setup lang="ts">
import { Search } from '@element-plus/icons-vue';
import dayjs from 'dayjs';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';

import { useShopStore } from '@/stores/useShopStore';
import type { MarketInput, MarketRecord } from '@/types';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const dialogVisible = ref(false);
const editingId = ref<number | null>(null);
const keyword = ref('');
const isMobile = ref(false);
const onlineIncomeDraft = ref(0);
const savingOnlineIncome = ref(false);

let mediaQuery: MediaQueryList | null = null;
const handleViewportChange = (event: MediaQueryListEvent) => {
  isMobile.value = event.matches;
};

const WEEKDAY_LABELS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

const emptyForm = (): MarketInput => ({
  name: '',
  date: dayjs().format('YYYY-MM-DD'),
  location: '',
  income: 0,
  note: '',
});

const form = reactive<MarketInput>(emptyForm());

const filteredMarkets = computed(() => {
  const kw = keyword.value.trim();

  if (!kw) return store.markets;

  return store.markets.filter((item) => item.name.includes(kw) || item.location.includes(kw));
});

const marketIncomeTotal = computed(() =>
  store.markets.reduce((sum, item) => sum + Number(item.income || 0), 0),
);

const totalIncome = computed(() => marketIncomeTotal.value + Number(store.onlineIncome || 0));

const totalExpense = computed(() =>
  store.expenses.reduce((sum, item) => sum + Number(item.amount || 0), 0),
);

const profit = computed(() => Number((totalIncome.value - totalExpense.value).toFixed(2)));

const isLoss = computed(() => profit.value < 0);

function formatProfit(value: number) {
  return value < 0 ? `-${formatCurrency(Math.abs(value))}` : formatCurrency(value);
}

const summaryCards = computed(() => [
  {
    label: '总收入',
    value: formatCurrency(totalIncome.value),
    hint: `集市 ${formatCurrency(marketIncomeTotal.value)} + 线上 ${formatCurrency(store.onlineIncome)}`,
    negative: false,
  },
  {
    label: '线上收入',
    value: formatCurrency(store.onlineIncome),
    hint: '手工维护，建议每周或每月更新一次',
    negative: false,
  },
  {
    label: '总支出',
    value: formatCurrency(totalExpense.value),
    hint: `来自支出记录 ${store.expenses.length} 条`,
    negative: false,
  },
  {
    label: '利润',
    value: formatProfit(profit.value),
    hint: isLoss.value ? `还差 ${formatCurrency(Math.abs(profit.value))} 才回本` : '总收入 − 总支出',
    negative: isLoss.value,
  },
]);

const onlineIncomeUpdatedText = computed(() =>
  store.onlineIncomeUpdatedAt
    ? dayjs(store.onlineIncomeUpdatedAt).format('YYYY-MM-DD HH:mm')
    : '还没有保存过',
);

watch(
  () => store.onlineIncome,
  (value) => {
    onlineIncomeDraft.value = Number(value) || 0;
  },
  { immediate: true },
);

function formatMarketDate(date: string) {
  const parsed = dayjs(date);

  return parsed.isValid() ? `${parsed.format('YYYY-MM-DD')}（${WEEKDAY_LABELS[parsed.day()]}）` : date;
}

function resetForm() {
  editingId.value = null;
  Object.assign(form, emptyForm());
}

function openCreateDialog() {
  resetForm();
  dialogVisible.value = true;
}

function openEditDialog(record: MarketRecord) {
  editingId.value = record.id ?? null;
  Object.assign(form, {
    name: record.name,
    date: record.date,
    location: record.location,
    income: record.income,
    note: record.note ?? '',
  });
  dialogVisible.value = true;
}

async function handleDeleteMarket(id: number) {
  try {
    await ElMessageBox.confirm('删除后这场集市的收入记录将无法恢复，确认继续吗？', '删除集市记录', {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'warning',
    });
    await store.deleteMarket(id);
    ElMessage.success('集市收入已删除');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '删除集市收入失败');
  }
}

async function submitMarket() {
  const name = form.name.trim();
  const location = form.location.trim();
  const income = Number(form.income);

  if (!name) {
    ElMessage.error('请填写集市名称');
    return;
  }

  if (!form.date) {
    ElMessage.error('请选择集市日期');
    return;
  }

  if (!location) {
    ElMessage.error('请填写集市地点');
    return;
  }

  if (!Number.isFinite(income) || income < 0) {
    ElMessage.error('收入需为不小于 0 的数字');
    return;
  }

  const payload: MarketInput = {
    name,
    date: form.date,
    location,
    income,
    note: form.note?.trim() ?? '',
  };

  try {
    if (editingId.value) {
      await store.updateMarket(editingId.value, payload);
      ElMessage.success('集市收入已更新');
    } else {
      await store.addMarket(payload);
      ElMessage.success('集市收入已保存');
    }

    dialogVisible.value = false;
    resetForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '集市收入保存失败');
  }
}

async function handleSaveOnlineIncome() {
  const value = Number(onlineIncomeDraft.value);

  if (!Number.isFinite(value) || value < 0) {
    ElMessage.error('线上收入需为不小于 0 的数字');
    return;
  }

  savingOnlineIncome.value = true;

  try {
    await store.saveOnlineIncome(value);
    ElMessage.success('线上收入已更新');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '线上收入保存失败');
  } finally {
    savingOnlineIncome.value = false;
  }
}

onMounted(() => {
  mediaQuery = window.matchMedia('(max-width: 768px)');
  isMobile.value = mediaQuery.matches;
  mediaQuery.addEventListener('change', handleViewportChange);
  store.initialize();
});

onUnmounted(() => {
  mediaQuery?.removeEventListener('change', handleViewportChange);
});
</script>

<template>
  <div class="page-stack">
    <section class="page-hero page-hero--warm">
      <div>
        <div class="hero-kicker">收入记录 · 线下集市 + 线上收入</div>
        <h2 class="hero-title">集市按场记、线上定期更新，收入支出一起算清楚</h2>
        <p class="hero-description">
          每次出摊结束后记一笔集市收入，线上收入按平台实际到账金额手工更新；页面会自动减去支出记录里的金额，算出当前利润。
        </p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">总收入 {{ formatCurrency(totalIncome) }}</span>
        <span class="hero-badge" :class="{ 'hero-badge--danger': isLoss }">
          利润 {{ formatProfit(profit) }}
        </span>
      </div>
    </section>

    <div class="summary-grid">
      <div
        v-for="item in summaryCards"
        :key="item.label"
        class="summary-tile"
        :class="{ 'summary-tile--negative': item.negative }"
      >
        <strong>{{ item.label }}</strong>
        <span>{{ item.value }}</span>
        <small class="summary-hint">{{ item.hint }}</small>
      </div>
    </div>

    <div v-if="isLoss" class="profit-warning">
      <strong>利润为负</strong>
      <span>
        当前总支出比总收入多 {{ formatCurrency(Math.abs(profit)) }}，请检查支出记录，或补录还没记的集市 / 线上收入。
      </span>
    </div>

    <el-card class="section-card">
      <div class="toolbar">
        <div>
          <h2>线上收入</h2>
          <div class="inline-tip">平台后台的实际到账金额，建议每周或每月手工更新一次，不用逐单记录</div>
        </div>
        <div class="online-actions">
          <el-input-number
            v-model="onlineIncomeDraft"
            class="online-input"
            :min="0"
            :step="100"
            :precision="2"
          />
          <el-button type="primary" :loading="savingOnlineIncome" @click="handleSaveOnlineIncome">
            保存
          </el-button>
        </div>
      </div>
      <div class="inline-tip">上次更新：{{ onlineIncomeUpdatedText }}</div>
    </el-card>

    <el-card class="section-card soft-table">
      <div class="toolbar">
        <div>
          <h2>集市收入台账</h2>
          <div class="inline-tip">
            共 {{ store.markets.length }} 场 · 累计 {{ formatCurrency(marketIncomeTotal) }}；线下出摊按场记录
          </div>
        </div>
        <div class="market-toolbar-actions">
          <el-input
            v-model="keyword"
            class="market-search"
            placeholder="搜索集市名称或地点"
            clearable
            :prefix-icon="Search"
          />
          <el-button type="primary" @click="openCreateDialog">新增集市</el-button>
        </div>
      </div>

      <div class="table-scroll">
        <el-table
          :data="filteredMarkets"
          :empty-text="keyword ? '没有匹配的集市记录，换个关键词试试。' : '还没有集市收入，先记下第一场集市吧。'"
        >
          <el-table-column prop="name" label="集市名称" min-width="160" />
          <el-table-column label="日期" width="160">
            <template #default="{ row }">{{ formatMarketDate(row.date) }}</template>
          </el-table-column>
          <el-table-column prop="location" label="地点" min-width="140" />
          <el-table-column label="收入" width="120">
            <template #default="{ row }">
              <span class="insight-value">{{ formatCurrency(row.income) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="备注" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">{{ row.note || '-' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openEditDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDeleteMarket(row.id!)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="editingId ? '编辑集市收入' : '新增集市收入'"
      :fullscreen="isMobile"
      width="640px"
      class="mobile-form-dialog"
    >
      <el-form :label-width="isMobile ? 'auto' : '92px'" :label-position="isMobile ? 'top' : 'right'">
        <div class="form-grid">
          <el-form-item label="集市名称">
            <el-input v-model="form.name" placeholder="例如：周末手帐市集" maxlength="30" show-word-limit />
          </el-form-item>
          <el-form-item label="日期">
            <el-date-picker
              v-model="form.date"
              type="date"
              value-format="YYYY-MM-DD"
              :clearable="false"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="地点">
            <el-input v-model="form.location" placeholder="例如：市中心文创园 A 区" maxlength="40" />
          </el-form-item>
          <el-form-item label="本场收入">
            <el-input-number v-model="form.income" :min="0" :step="10" :precision="2" />
          </el-form-item>
          <el-form-item label="备注" class="form-grid-wide">
            <el-input
              v-model="form.note"
              type="textarea"
              :rows="3"
              maxlength="200"
              show-word-limit
              placeholder="可记录摊位费、客流、卖得好的商品等"
            />
          </el-form-item>
        </div>
      </el-form>

      <div class="dialog-tip">一场集市记一条，日期按实际出摊当天填写。收入填写当天实际到手的金额即可。</div>

      <template #footer>
        <div class="dialog-actions">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitMarket">{{ editingId ? '更新记录' : '保存记录' }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.summary-hint {
  display: block;
  margin-top: 6px;
  color: #8a93a8;
  font-size: 12px;
  line-height: 1.5;
}

.summary-tile--negative {
  border-color: rgba(229, 72, 77, 0.32);
  background: #fff7f7;
}

.summary-tile--negative span {
  color: #d92d36;
}

.summary-tile--negative .summary-hint {
  color: #c1656a;
}

.hero-badge--danger {
  background: rgba(229, 72, 77, 0.14);
  color: #b42318;
}

.profit-warning {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
  padding: 14px 18px;
  border-radius: 16px;
  border: 1px solid rgba(229, 72, 77, 0.3);
  background: #fff1f1;
  color: #b42318;
  font-size: 13px;
}

.profit-warning strong {
  font-size: 14px;
}

.online-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.online-input {
  width: 220px;
}

.market-toolbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.market-search {
  width: 240px;
}

.form-grid-wide {
  grid-column: 1 / -1;
}

@media (max-width: 768px) {
  .online-actions {
    flex-direction: column;
    align-items: stretch;
    width: 100%;
  }

  .online-input {
    width: 100%;
  }

  .market-toolbar-actions {
    flex-direction: column;
    align-items: stretch;
    width: 100%;
  }

  .market-search {
    width: 100%;
  }
}
</style>
