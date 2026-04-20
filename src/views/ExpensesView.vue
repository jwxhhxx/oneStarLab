<script setup lang="ts">
import dayjs from 'dayjs';
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';

import { useShopStore } from '@/stores/useShopStore';
import type { ExpenseInput, ExpenseRecord } from '@/types';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const dialogVisible = ref(false);
const editingExpenseId = ref<number | null>(null);
const isMobile = ref(false);

let mediaQuery: MediaQueryList | null = null;
const handleViewportChange = (event: MediaQueryListEvent) => {
  isMobile.value = event.matches;
};

const emptyForm = (): ExpenseInput => ({
  purpose: '',
  amount: 0,
});

const form = reactive<ExpenseInput>(emptyForm());

const summaryCards = computed(() => {
  const totalExpense = store.expenses.reduce((sum, item) => sum + item.amount, 0);
  const currentMonthExpense = store.expenses
    .filter((item) => dayjs(item.createdAt).isSame(dayjs(), 'month'))
    .reduce((sum, item) => sum + item.amount, 0);

  return [
    { label: '累计支出笔数', value: `${store.expenses.length}` },
    { label: '累计支出金额', value: formatCurrency(totalExpense) },
    { label: '本月支出', value: formatCurrency(currentMonthExpense) },
    { label: '最近记录', value: store.expenses[0] ? dayjs(store.expenses[0].createdAt).format('MM-DD HH:mm') : '暂无' },
  ];
});

function resetForm() {
  editingExpenseId.value = null;
  Object.assign(form, emptyForm());
}

function openCreateDialog() {
  resetForm();
  dialogVisible.value = true;
}

function openEditDialog(item: ExpenseRecord) {
  editingExpenseId.value = item.id ?? null;
  Object.assign(form, {
    purpose: item.purpose,
    amount: item.amount,
  });
  dialogVisible.value = true;
}

async function submitExpense() {
  try {
    if (!form.purpose.trim()) {
      throw new Error('请填写支出用途');
    }

    if (form.amount <= 0) {
      throw new Error('支出金额必须大于 0');
    }

    if (editingExpenseId.value) {
      await store.updateExpense(editingExpenseId.value, { ...form });
      ElMessage.success('支出记录已更新');
    } else {
      await store.addExpense({ ...form });
      ElMessage.success('支出记录已添加');
    }

    dialogVisible.value = false;
    resetForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存支出记录失败');
  }
}

async function handleDeleteExpense(id: number) {
  try {
    await store.deleteExpense(id);
    ElMessage.success('支出记录已删除');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除支出记录失败');
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
        <div class="hero-kicker">支出记录 · 店铺成本台账</div>
        <h2 class="hero-title">记录开店以来每一笔实际支出</h2>
        <p class="hero-description">把包装、设备、样品、运费和日常运营花费都集中管理，后续也便于核算真实投入。</p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">累计 {{ store.expenses.length }} 笔</span>
        <span class="hero-badge">总支出 {{ formatCurrency(store.expenses.reduce((sum, item) => sum + item.amount, 0)) }}</span>
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
          <h2>支出记录</h2>
          <div class="inline-tip">记录用途和金额，形成完整的开店支出台账</div>
        </div>
        <el-button type="primary" @click="openCreateDialog">新增支出</el-button>
      </div>

      <div class="table-scroll">
        <el-table :data="store.expenses" empty-text="还没有支出记录，先录入一笔。">
          <el-table-column label="记录时间" min-width="160">
            <template #default="{ row }">{{ dayjs(row.createdAt).format('YYYY-MM-DD HH:mm') }}</template>
          </el-table-column>
          <el-table-column prop="purpose" label="支出用途" min-width="220" />
          <el-table-column label="金额" width="140">
            <template #default="{ row }">
              <span class="insight-value">{{ formatCurrency(row.amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openEditDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDeleteExpense(row.id!)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="editingExpenseId ? '编辑支出记录' : '新增支出记录'"
      :fullscreen="isMobile"
      width="640px"
      class="mobile-form-dialog"
    >
      <el-form :label-width="isMobile ? 'auto' : '92px'" :label-position="isMobile ? 'top' : 'right'">
        <div class="form-grid">
          <el-form-item label="支出用途">
            <el-input v-model="form.purpose" placeholder="例如：包装材料、样品打样、设备购置" />
          </el-form-item>
          <el-form-item label="支出金额">
            <el-input-number v-model="form.amount" :min="0" :step="0.1" />
          </el-form-item>
        </div>
      </el-form>

      <div class="dialog-tip">建议按真实支出逐笔登记，后续更方便复盘开店总投入。</div>

      <template #footer>
        <div class="dialog-actions">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitExpense">{{ editingExpenseId ? '更新记录' : '保存记录' }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
