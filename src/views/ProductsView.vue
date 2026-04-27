<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { ElMessage } from 'element-plus';
import JsBarcode from 'jsbarcode';

import { useShopStore } from '@/stores/useShopStore';
import type { Product, ProductInput } from '@/types';
import { formatCurrency } from '@/utils/pricing';

const store = useShopStore();
const dialogVisible = ref(false);
const editingProductId = ref<number | null>(null);
const isMobile = ref(false);

let mediaQuery: MediaQueryList | null = null;
const handleViewportChange = (event: MediaQueryListEvent) => {
  isMobile.value = event.matches;
};

const emptyForm = (): ProductInput => ({
  name: '',
  sku: '',
  category: store.categories.length ? store.categories[0].name : '胶带',
  supplier: '',
  purchaseCost: 0,
  packagingCost: 0,
  salePrice: 0,
  stock: 0,
  safeStock: 5,
  note: '',
});

const form = reactive<ProductInput>(emptyForm());
const newCategoryName = ref('');

const tableRows = computed(() =>
  store.products.map((item) => ({
    ...item,
    suggestedPrice: store.getSuggestedPrice(item),
    isLowStock: item.stock <= item.safeStock,
  })),
);

const summaryCards = computed(() => {
  const stockValue = store.products.reduce(
    (sum, item) => sum + (item.purchaseCost + item.packagingCost) * item.stock,
    0,
  );

  return [
    { label: '商品总数', value: `${store.products.length}` },
    { label: '低库存商品', value: `${tableRows.value.filter((item) => item.isLowStock).length}` },
    { label: '库存成本估值', value: formatCurrency(stockValue) },
    { label: '平均建议售价', value: formatCurrency(tableRows.value.length ? tableRows.value.reduce((sum, item) => sum + item.suggestedPrice, 0) / tableRows.value.length : 0) },
  ];
});

function resetForm() {
  editingProductId.value = null;
  Object.assign(form, emptyForm());
}

async function addCategoryUI() {
  const nm = (newCategoryName.value || '').trim();
  if (!nm) return;
  await store.addCategory(nm);
  newCategoryName.value = '';
}

async function handleDeleteCategory(id?: number) {
  if (!id) return;
  if (!confirm('确认删除该分类？已被使用的商品不会被修改')) return;
  await store.deleteCategory(id);
}

function openCreateDialog() {
  resetForm();
  dialogVisible.value = true;
}

function openEditDialog(item: Product) {
  editingProductId.value = item.id ?? null;
  Object.assign(form, {
    name: item.name,
    sku: item.sku,
    category: item.category,
    supplier: item.supplier,
    purchaseCost: item.purchaseCost,
    packagingCost: item.packagingCost,
    salePrice: item.salePrice,
    stock: item.stock,
    safeStock: item.safeStock,
    note: item.note ?? '',
  });
  dialogVisible.value = true;
}

async function handleDeleteProduct(id: number) {
  try {
    await store.deleteProduct(id);
    ElMessage.success('商品已删除');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除失败');
  }
}

function escapeHtml(str: string) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function printProductBarcode(product: Product) {
  const value = product.sku || product.name;
  if (!value) {
    window.alert('无可用 SKU 或条码');
    return;
  }

  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  try {
    JsBarcode(svg as any, value, { format: 'CODE128', displayValue: true, fontSize: 14, height: 60, width: 2 } as any);
  } catch (e) {
    console.error(e);
    window.alert('生成条码失败');
    return;
  }

  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(svg);
  const html = `
  <html>
    <head>
      <title>打印条码 - ${escapeHtml(product.name)}</title>
      <style>
        @page { size: 38mm 25mm; margin: 0; }
        body { margin: 0; display:flex; align-items:center; justify-content:center; height:100vh; }
        .label { width:38mm; height:25mm; display:flex; flex-direction:column; align-items:center; justify-content:center; }
        .name { font-size:12px; margin-top:4px; }
      </style>
    </head>
    <body>
      <div class="label">${svgString}<div class="name">${escapeHtml(product.name)}</div></div>
      <script>window.onload = () => { window.print(); setTimeout(()=>window.close(), 200); };<\/script>
    </body>
  </html>`;

  const w = window.open('', '_blank');
  if (w) {
    w.document.write(html);
    w.document.close();
  } else {
    window.alert('浏览器阻止了弹窗，请允许弹窗以打印条码');
  }
}

async function submitProduct() {
  try {
    if (editingProductId.value) {
      await store.updateProduct(editingProductId.value, { ...form });
      ElMessage.success('商品已更新');
    } else {
      await store.addProduct({ ...form });
      ElMessage.success('商品已录入');
    }

    dialogVisible.value = false;
    resetForm();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败');
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
        <div class="hero-kicker">商品中心 · 更适合日常录入</div>
        <h2 class="hero-title">把库存、成本和建议售价放在同一视图里</h2>
        <p class="hero-description">更像经营台账的视觉风格，新增商品和查看库存都更直观。</p>
      </div>
      <div class="hero-badges">
        <span class="hero-badge">在售 {{ store.products.length }} 款</span>
        <span class="hero-badge">预警 {{ tableRows.filter((item) => item.isLowStock).length }} 款</span>
      </div>
    </section>

    <div class="summary-grid">
      <div v-for="item in summaryCards" :key="item.label" class="summary-tile">
        <strong>{{ item.label }}</strong>
        <span>{{ item.value }}</span>
      </div>
    </div>

    <el-card class="section-card" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap;">
        <div style="flex:1">
          <strong>商品分类</strong>
          <div class="inline-tip">新增的分类会自动出现在新增/编辑商品的下拉中。</div>
        </div>
        <el-input v-model="newCategoryName" placeholder="新增分类" style="width:220px" @keyup.enter.native="addCategoryUI" />
        <el-button type="primary" @click="addCategoryUI">添加分类</el-button>
      </div>

      <div style="margin-top:12px">
        <el-tag v-for="c in store.categories" :key="c.id" closable @close="handleDeleteCategory(c.id)" style="margin-right:8px;margin-bottom:8px;">
          {{ c.name }}
        </el-tag>
        <div v-if="!store.categories.length" style="color:#888;margin-top:8px">未定义分类，新增商品时可直接输入分类并保存。</div>
      </div>
    </el-card>

    <el-card class="section-card soft-table">
      <div class="toolbar">
        <div>
          <h2>商品中心</h2>
          <div class="inline-tip">录入商品、查看库存、参考建议售价</div>
        </div>
        <el-button type="primary" @click="openCreateDialog">新增商品</el-button>
      </div>

      <div class="table-scroll">
        <el-table :data="tableRows" empty-text="暂无商品，先新增一条吧。">
          <el-table-column prop="name" label="商品名称" min-width="220" />
          <el-table-column label="分类" width="100">
            <template #default="{ row }">
              <el-tag effect="plain" round>{{ row.category }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sku" label="SKU" width="140" />
          <el-table-column label="成本" width="120">
            <template #default="{ row }">{{ formatCurrency(row.purchaseCost + row.packagingCost) }}</template>
          </el-table-column>
          <el-table-column label="当前售价" width="120">
            <template #default="{ row }">{{ formatCurrency(row.salePrice) }}</template>
          </el-table-column>
          <el-table-column label="建议售价" width="120">
            <template #default="{ row }">
              <span class="insight-value">{{ formatCurrency(row.suggestedPrice) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="库存" width="110">
            <template #default="{ row }">
              <el-tag :type="row.isLowStock ? 'danger' : 'success'" round>{{ row.stock }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="supplier" label="供应商" min-width="120" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openEditDialog(row)">编辑</el-button>
              <el-button link type="info" @click="printProductBarcode(row)">打印条形码</el-button>
              <el-button link type="danger" @click="handleDeleteProduct(row.id!)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="editingProductId ? '编辑商品' : '新增商品'"
      :fullscreen="isMobile"
      width="720px"
      class="mobile-form-dialog"
    >
      <el-form :label-width="isMobile ? 'auto' : '92px'" :label-position="isMobile ? 'top' : 'right'">
        <div class="form-grid">
          <el-form-item label="商品名称"><el-input v-model="form.name" /></el-form-item>
          <el-form-item label="SKU"><el-input v-model="form.sku" /></el-form-item>
          <el-form-item label="分类">
            <el-select v-model="form.category" filterable allow-create placeholder="选择或输入分类">
              <el-option v-for="c in store.categories" :key="c.id" :label="c.name" :value="c.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="供应商"><el-input v-model="form.supplier" /></el-form-item>
          <el-form-item label="采购价"><el-input-number v-model="form.purchaseCost" :min="0" :step="0.1" /></el-form-item>
          <el-form-item label="包装成本"><el-input-number v-model="form.packagingCost" :min="0" :step="0.1" /></el-form-item>
          <el-form-item label="当前售价"><el-input-number v-model="form.salePrice" :min="0" :step="0.1" /></el-form-item>
          <el-form-item label="库存"><el-input-number v-model="form.stock" :min="0" /></el-form-item>
          <el-form-item label="安全库存"><el-input-number v-model="form.safeStock" :min="0" /></el-form-item>
          <el-form-item label="备注"><el-input v-model="form.note" /></el-form-item>
        </div>
      </el-form>

      <div class="dialog-tip">如果不填写售价，系统会自动按当前定价规则给出建议值。</div>

      <template #footer>
        <div class="dialog-actions">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitProduct">{{ editingProductId ? '更新商品' : '保存商品' }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>