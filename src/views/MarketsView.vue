<script setup lang="ts">
import { Search } from '@element-plus/icons-vue';
import dayjs from 'dayjs';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';

import { useShopStore } from '@/stores/useShopStore';
import type { MarketInput, MarketRecord } from '@/types';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const dialogVisible = ref(false);
const editingId = ref<number | null>(null);
const keyword = ref('');
const isMobile = ref(false);

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

const totalIncome = computed(() => store.markets.reduce((sum, item) => sum + item.income, 0));

const summaryCards = computed(() => {
  const currentMonth = dayjs().format('YYYY-MM');
  const monthIncome = store.markets
    .filter((item) => item.date.startsWith(currentMonth))
    .reduce((sum, item) => sum + item.income, 0);

  return [
    { label: '累计场次', value: `${store.markets.length} 场` },
    { label: '累计收入', value: formatCurrency(totalIncome.value) },
    { label: '本月收入', value: formatCurrency(monthIncome) },
    {
      label: '场均收入',
      value: formatCurrency(store.markets.length ? totalIncome.value / store.markets.length : 0),
    },
  ];
});

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
    ElMessage.success('集市记录已删除');
  } catch (error) {
    if (error === 'cancel' || error === 'close') return;
    ElMessage.error(error instanceof Error ? error.message : '删除集市记录失败');
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
      ElMessage.success('集市记录已更新');
    } else {
      await store.addMarket(payload);
      ElMessage.success('集市记录已保存');
    }

    dialogVisible.value = false;
    resetForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '集市记录保存失败');
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
        <div class="hero-kicker">集市记录 · 线下出摊台账</div>
        <h2 class="hero-title">每一场集市的时间、地点和收入都留在这里</h2>
        <p class="hero-description">出摊结束后顺手记一笔，回头就能看出哪类市集更值得去、一个月摆摊能赚多少。</p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">累计 {{ store.markets.length }} 场</span>
        <span class="hero-badge">总收入 {{ formatCurrency(totalIncome) }}</span>
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
          <h2>集市记录</h2>
          <div class="inline-tip">只统计线下出摊收入，不计入线上订单的销售额与利润</div>
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
          :empty-text="keyword ? '没有匹配的集市记录，换个关键词试试。' : '还没有集市记录，先记下第一场集市吧。'"
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
      :title="editingId ? '编辑集市记录' : '新增集市记录'"
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
