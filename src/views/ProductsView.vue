<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, nextTick, watch } from 'vue';
import { ElMessage } from 'element-plus';

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

const categoryCounts = computed(() => store.getCategoryUsageCounts());

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
  const cat = store.categories.find((x) => x.id === id);
  if (!cat) return;
  const count = categoryCounts.value[cat.name] || 0;
  if (count > 0) {
    const other = store.categories.filter((x) => x.id !== id).map((x) => x.name).join('，') || '无';
    const msg = `该分类下有 ${count} 件商品。输入目标分类名称以重分配（或留空清空分类），取消则放弃。可选目标：${other}`;
    const input = window.prompt(msg, other === '无' ? '' : '');
    if (input === null) return; // canceled
    const target = (input || '').trim();
    await store.deleteCategory(id, target || undefined);
  } else {
    if (!confirm('确认删除该分类？')) return;
    await store.deleteCategory(id);
  }
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

// Barcode modal state and helpers
const barcodeModalVisible = ref(false);
const barcodeModalProduct = ref<Product | null>(null);
const barcodeModalValue = ref('');
const barcodeSvgRef = ref<SVGSVGElement | null>(null);
const barcodePreviewDataUrl = ref<string | null>(null);
// barcode options
const barcodeLabelSize = ref('38x25');
const barcodeOrientation = ref<'portrait' | 'landscape'>('landscape');
const barcodeFontSize = ref(14);

const labelSizeOptions = [
  { value: '38x25', label: '38×25 mm' },
  { value: '50x30', label: '50×30 mm' },
];

let _jsBarcode: any = null;
async function ensureJsBarcode() {
  if (_jsBarcode) return _jsBarcode;
  const mod = await import('jsbarcode');
  _jsBarcode = mod.default ?? mod;
  return _jsBarcode;
}

let regenTimer: any = null;

function sanitizeFileName(name: string) {
  if (!name) return '';
  return name.replace(/[\\/:*?"<>|]/g, '').trim();
}

function isSkuValid(sku?: string) {
  if (!sku) return false;
  return /^\d+-\d{4,}$/.test(sku);
}

async function calibrateSku() {
  try {
    const target = await store.generateSkuForCategory(form.category, editingProductId?.value ?? undefined);
    form.sku = target;
    ElMessage.success('已校准 SKU: ' + target);
  } catch (e) {
    console.error(e);
    ElMessage.error('校准失败');
  }
}

async function openBarcodeModal(product: Product) {
  barcodeModalProduct.value = product;
  barcodeModalValue.value = product.sku || product.name || '';
  barcodePreviewDataUrl.value = null;
  barcodeSvgRef.value = null;
  barcodeModalVisible.value = true;
  await nextTick();
  generateBarcode();
}

async function generateBarcode(silent = false) {
  const value = barcodeModalValue.value || barcodeModalProduct.value?.sku || barcodeModalProduct.value?.name || '';
  if (!value) {
    if (!silent) window.alert('请输入条码或 SKU');
    return;
  }

  const heightPx = barcodeLabelSize.value === '38x25' ? 60 : 80;
  const opts = { format: 'CODE128', displayValue: true, fontSize: barcodeFontSize.value, height: heightPx, width: 2 } as any;

  let svgEl = barcodeSvgRef.value as any;
  if (!svgEl) {
    const created = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    barcodeSvgRef.value = created as unknown as SVGSVGElement;
    svgEl = created as any;
  }

  try {
    const JsBarcode = await ensureJsBarcode();
    JsBarcode(svgEl, value, opts);
  } catch (e) {
    console.error(e);
    window.alert('生成条码失败');
    return;
  }

  try {
    const canvas = document.createElement('canvas');
    const JsBarcodeLib = await ensureJsBarcode();
    JsBarcodeLib(canvas as any, value, opts);
    const out = document.createElement('canvas');
    // handle orientation: portrait rotates the barcode
    if (barcodeOrientation.value === 'portrait') {
      out.width = canvas.height || 80;
      out.height = canvas.width || 300;
    } else {
      out.width = canvas.width || 300;
      out.height = canvas.height || 80;
    }
    const ctx = out.getContext('2d');
    if (ctx) {
      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, out.width, out.height);
      if (barcodeOrientation.value === 'portrait') {
        ctx.save();
        ctx.translate(out.width / 2, out.height / 2);
        ctx.rotate(-Math.PI / 2);
        ctx.drawImage(canvas, -canvas.width / 2, -canvas.height / 2);
        ctx.restore();
      } else {
        ctx.drawImage(canvas, 0, 0);
      }
      barcodePreviewDataUrl.value = out.toDataURL('image/png');
    } else {
      barcodePreviewDataUrl.value = null;
    }
  } catch (e) {
    console.warn('生成预览失败', e);
    barcodePreviewDataUrl.value = null;
  }
}

// 自动更新预览：当方向、标签尺寸、字体或输入值变化时，若弹窗打开则防抖触发生成
watch(
  [barcodeOrientation, barcodeLabelSize, barcodeFontSize, barcodeModalValue, barcodeModalVisible],
  () => {
    if (!barcodeModalVisible.value) return;
    if (regenTimer) clearTimeout(regenTimer);
    regenTimer = setTimeout(() => {
      generateBarcode(true);
    }, 150);
  },
);

async function downloadBarcodePNG() {
  // ensure fresh preview synchronously generated
  const value = barcodeModalValue.value || barcodeModalProduct.value?.sku || barcodeModalProduct.value?.name || '';
  if (!value) {
    window.alert('请输入条码或 SKU');
    return;
  }

  // regenerate synchronously to avoid async-share issues on some mobile browsers
  try {
    const JsBarcodeLib = await ensureJsBarcode();
    const canvas = document.createElement('canvas');
    const heightPx = barcodeLabelSize.value === '38x25' ? 60 : 80;
    const opts = { format: 'CODE128', displayValue: true, fontSize: barcodeFontSize.value, height: heightPx, width: 2 } as any;
    JsBarcodeLib(canvas as any, value, opts);

    const out = document.createElement('canvas');
    if (barcodeOrientation.value === 'portrait') {
      out.width = canvas.height || 80;
      out.height = canvas.width || 300;
    } else {
      out.width = canvas.width || 300;
      out.height = canvas.height || 80;
    }
    const ctx = out.getContext('2d');
    if (ctx) {
      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, out.width, out.height);
      if (barcodeOrientation.value === 'portrait') {
        ctx.save();
        ctx.translate(out.width / 2, out.height / 2);
        ctx.rotate(-Math.PI / 2);
        ctx.drawImage(canvas, -canvas.width / 2, -canvas.height / 2);
        ctx.restore();
      } else {
        ctx.drawImage(canvas, 0, 0);
      }
      const dataUrl = out.toDataURL('image/png');

      const selected = barcodeModalProduct.value;
      const base = selected ? selected.name : 'label';
      const safe = sanitizeFileName(base) || 'label';
      const time = new Date().toISOString().replace(/[:.]/g, '-');
      const fileName = `${safe}-${time}.png`;

      const a = document.createElement('a');
      a.href = dataUrl;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      a.remove();
      // also update preview
      barcodePreviewDataUrl.value = dataUrl;
    } else {
      window.alert('生成图片失败');
    }
  } catch (e) {
    console.error('下载失败', e);
    window.alert('下载失败');
  }
}

async function printBarcode() {
  const value = barcodeModalValue.value || barcodeModalProduct.value?.sku || barcodeModalProduct.value?.name || '';
  if (!value) {
    window.alert('请输入条码或 SKU');
    return;
  }

  // generate image respecting orientation
  try {
    const JsBarcodeLib = await ensureJsBarcode();
    const canvas = document.createElement('canvas');
    const heightPx = barcodeLabelSize.value === '38x25' ? 60 : 80;
    const opts = { format: 'CODE128', displayValue: true, fontSize: barcodeFontSize.value, height: heightPx, width: 2 } as any;
    JsBarcodeLib(canvas as any, value, opts);

    const out = document.createElement('canvas');
    if (barcodeOrientation.value === 'portrait') {
      out.width = canvas.height || 80;
      out.height = canvas.width || 300;
    } else {
      out.width = canvas.width || 300;
      out.height = canvas.height || 80;
    }
    const ctx = out.getContext('2d');
    if (!ctx) throw new Error('无法获得画布上下文');
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, out.width, out.height);
    if (barcodeOrientation.value === 'portrait') {
      ctx.save();
      ctx.translate(out.width / 2, out.height / 2);
      ctx.rotate(-Math.PI / 2);
      ctx.drawImage(canvas, -canvas.width / 2, -canvas.height / 2);
      ctx.restore();
    } else {
      ctx.drawImage(canvas, 0, 0);
    }

    const dataUrl = out.toDataURL('image/png');

    // determine label size in mm for @page
    let w = 38;
    let h = 25;
    if (barcodeLabelSize.value === '50x30') {
      w = 50;
      h = 30;
    }
    if (barcodeOrientation.value === 'portrait') {
      const tmp = w; w = h; h = tmp;
    }

    const productName = barcodeModalProduct.value?.name ?? '';
    const html = `
      <html>
        <head>
          <title>打印条码 - ${escapeHtml(productName)}</title>
          <style>
            @page { size: ${w}mm ${h}mm; margin: 0; }
            body { margin: 0; display:flex; align-items:center; justify-content:center; height:100vh; }
            .label { width:${w}mm; height:${h}mm; display:flex; flex-direction:column; align-items:center; justify-content:center; }
            .label img{ max-width:100%; max-height:100%; display:block; }
            .name { font-size:12px; margin-top:4px; }
          </style>
        </head>
        <body>
          <div class="label"><img src="${dataUrl}" alt="barcode" /><div class="name">${escapeHtml(productName)}</div></div>
          <script>window.onload = () => { window.print(); setTimeout(()=>window.close(), 200); };<\/script>
        </body>
      </html>`;

    const win = window.open('', '_blank');
    if (win) {
      win.document.write(html);
      win.document.close();
    } else {
      window.alert('浏览器阻止了弹窗，请允许弹窗以打印条码');
    }
  } catch (e) {
    console.error('打印失败', e);
    window.alert('打印失败');
  }
}

function closeBarcodeModal() {
  barcodeModalVisible.value = false;
  barcodeModalProduct.value = null;
  barcodeModalValue.value = '';
  barcodePreviewDataUrl.value = null;
  barcodeSvgRef.value = null;
}

function copySku() {
  const txt = barcodeModalValue.value || '';
  if (!txt) {
    ElMessage.warning('当前无可复制的 SKU');
    return;
  }
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(txt).then(() => ElMessage.success('已复制 SKU'), () => ElMessage.error('复制失败'));
  } else {
    const ta = document.createElement('textarea');
    ta.value = txt;
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand('copy');
      ElMessage.success('已复制 SKU');
    } catch (e) {
      ElMessage.error('复制失败');
    } finally {
      ta.remove();
    }
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
        <div class="category-controls" style="align-items:center;">
          <div style="flex:1">
            <strong>商品分类</strong>
            <div class="inline-tip">新增的分类会自动出现在新增/编辑商品的下拉中。</div>
          </div>
          <div class="controls-right">
            <el-input v-model="newCategoryName" placeholder="新增分类" class="add-input" @keyup.enter.native="addCategoryUI" />
            <el-button type="primary" class="add-button" @click="addCategoryUI">添加分类</el-button>
          </div>
        </div>

        <div class="categories-section">
          <div v-if="!store.categories.length" class="empty-cat" style="color:#888;margin-top:8px">未定义分类，新增商品时可直接输入分类并保存。</div>
          <div v-else class="categories-grid">
            <el-tag v-for="c in store.categories" :key="c.id" closable @close="handleDeleteCategory(c.id)" class="category-tag">
              <span class="cat-label">{{ c.name }}</span>
              <span class="cat-count">· {{ categoryCounts[c.name] || 0 }}</span>
            </el-tag>
          </div>
        </div>
    </el-card>

    <el-card class="section-card soft-table">
      <div class="toolbar">
        <div>
          <h2>商品列表</h2>
          <div class="inline-tip">录入商品、查看库存、参考建议售价</div>
        </div>
        <el-button type="primary" @click="openCreateDialog">新增商品</el-button>
      </div>

      <div class="table-scroll">
        <el-table :data="tableRows" empty-text="暂无商品，先新增一条吧。">
          <el-table-column prop="name" label="商品名称" min-width="160" show-overflow-tooltip />
          <el-table-column label="分类" width="100">
            <template #default="{ row }">
              <el-tag effect="plain" round>{{ row.category }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="SKU" width="160">
            <template #default="{ row }">
              <span :class="['sku-text', !isSkuValid(row.sku) ? 'sku-invalid' : '']">{{ row.sku || '—' }}</span>
              <span v-if="!isSkuValid(row.sku)" style="margin-left:6px;color:#e6a23c;font-size:12px">(格式异常)</span>
            </template>
          </el-table-column>
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
              <el-button link type="info" @click="openBarcodeModal(row)">条码</el-button>
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
          <el-form-item label="SKU">
            <div style="display:flex;gap:8px;align-items:center">
              <el-input v-model="form.sku" placeholder="自动生成 SKU" :disabled="!editingProductId" />
              <el-button type="text" v-if="editingProductId" @click="calibrateSku">校准</el-button>
            </div>
          </el-form-item>
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

    <el-drawer v-if="isMobile" v-model="barcodeModalVisible" :title="barcodeModalProduct ? '条形码 - ' + barcodeModalProduct.name : '条形码'" direction="btt" size="50%" :destroy-on-close="true">
      <div style="display:flex;gap:8px;align-items:center;margin-bottom:8px;">
        <el-input v-model="barcodeModalValue" placeholder="条码或 SKU" style="flex:1" :disabled="true" />
        <el-button type="primary" @click="generateBarcode">生成</el-button>
        <el-button type="text" @click="copySku">复制</el-button>
      </div>

      <div style="display:flex;gap:12px;flex-wrap:wrap;align-items:center;margin-bottom:8px;">
        <el-select v-model="barcodeLabelSize" placeholder="标签尺寸" style="width:140px">
          <el-option v-for="opt in labelSizeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
        <el-radio-group v-model="barcodeOrientation">
          <el-radio-button label="landscape">横向</el-radio-button>
          <el-radio-button label="portrait">纵向</el-radio-button>
        </el-radio-group>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:12px;color:#666">字体</span>
          <el-slider v-model="barcodeFontSize" :min="8" :max="24" style="width:120px" />
        </div>
      </div>

      <div style="margin-top:6px">
        <div v-if="barcodePreviewDataUrl">
          <img :src="barcodePreviewDataUrl" alt="barcode" style="width:100%;max-height:160px;background:#fff;padding:6px;border-radius:6px;object-fit:contain;" />
          <div style="margin-top:8px;display:flex;gap:8px">
            <el-button @click="downloadBarcodePNG">下载 PNG</el-button>
            <el-button type="success" @click="printBarcode">打印</el-button>
          </div>
          <div style="font-size:12px;color:#888;margin-top:6px">长按图片也可保存到手机（iOS/部分 WebView 需长按）</div>
        </div>
        <div v-else style="color:#888">请点击“生成”以预览条码。</div>
      </div>

      <template #footer>
        <el-button @click="closeBarcodeModal">关闭</el-button>
      </template>
    </el-drawer>

    <el-dialog v-else v-model="barcodeModalVisible" :title="barcodeModalProduct ? '条形码 - ' + barcodeModalProduct.name : '条形码'" width="520px" :destroy-on-close="true">
      <div style="display:flex;gap:8px;align-items:center;">
        <el-input v-model="barcodeModalValue" placeholder="条码或 SKU" style="flex:1" :disabled="true" />
        <el-button type="primary" @click="generateBarcode">生成</el-button>
        <el-button type="text" @click="copySku">复制</el-button>
      </div>

      <div style="display:flex;gap:12px;flex-wrap:wrap;align-items:center;margin-top:12px;">
        <el-select v-model="barcodeLabelSize" placeholder="标签尺寸" style="width:160px">
          <el-option v-for="opt in labelSizeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
        <el-radio-group v-model="barcodeOrientation">
          <el-radio-button label="landscape">横向</el-radio-button>
          <el-radio-button label="portrait">纵向</el-radio-button>
        </el-radio-group>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:12px;color:#666">字体</span>
          <el-slider v-model="barcodeFontSize" :min="8" :max="24" style="width:160px" />
        </div>
      </div>

      <div style="margin-top:12px">
        <div v-if="barcodePreviewDataUrl">
          <img :src="barcodePreviewDataUrl" alt="barcode" style="height:120px;background:#fff;padding:6px;border-radius:6px;" />
          <div style="margin-top:8px;display:flex;gap:8px">
            <el-button @click="downloadBarcodePNG">下载 PNG</el-button>
            <el-button type="success" @click="printBarcode">打印</el-button>
          </div>
          <div style="font-size:12px;color:#888;margin-top:6px">长按图片也可保存到手机</div>
        </div>
        <div v-else style="color:#888">请点击“生成”以预览条码。</div>
      </div>

      <template #footer>
        <el-button @click="closeBarcodeModal">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.category-controls{display:flex;gap:12px;flex-wrap:wrap}
.controls-right{display:flex;gap:8px;align-items:center}
.add-input{width:220px}
.add-button{white-space:nowrap}
.categories-section{margin-top:12px}
.categories-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(120px,1fr));gap:8px;align-items:start}
.category-tag{display:flex;justify-content:space-between;align-items:center;padding:4px 8px}
.cat-count{opacity:0.7;margin-left:6px;font-size:12px}
.empty-cat{color:#888;margin-top:8px}

.sku-text{display:inline-block;max-width:120px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.sku-invalid{color:#e6a23c;font-weight:600}

@media (max-width: 768px){
  .category-controls{flex-direction:column;align-items:stretch}
  .controls-right{flex-direction:row;gap:8px}
  .add-input{width:100%}
  .add-button{width:100%}
  .category-tag{padding:8px}
  .categories-grid{grid-template-columns:repeat(auto-fit,minmax(140px,1fr))}
}

@media (min-width: 769px){
  .categories-grid{display:flex;flex-wrap:wrap}
  .category-tag{margin-right:8px;margin-bottom:8px}
}
</style>